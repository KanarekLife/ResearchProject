package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"mcbench/game"
)

// Answer protocols.
const (
	// ProtocolTools: the model calls the game tools itself (function calling).
	ProtocolTools = "tools"
	// ProtocolJSON: for models without tool calling. The harness calls
	// get_state and get_decision for the model and puts the results in the
	// prompt; the model answers {"reasoning": ..., "option_id": N}, which
	// the harness passes to choose_option.
	ProtocolJSON = "json"
)

// LLM plays through an OpenAI-compatible chat API (local or hosted).
//
// A game is one append-only conversation (the agentic loop): each decision
// adds a user message with the events since the last decision and the
// current decision, then the model's replies (with their reasoning) and any
// tool results. Nothing is ever removed or rewritten, so each request's
// prompt is a prefix of the next and the server's prompt cache stays valid.
type LLM struct {
	Chat         Chat
	Protocol     string
	MaxSteps     int // model requests allowed per decision
	Instructions Instructions
}

func (l *LLM) Name() string {
	return fmt.Sprintf("llm:%s:%s:t%g", l.Chat.Model, l.Protocol, l.Chat.Temperature)
}

const harnessPrompt = `You are playing a solo game of Marvel Champions: The Card Game as the hero player, from setup until the game ends.
The game engine enforces all rules: only the options it lists are legal.`

const toolsPrompt = `Play using the tools. get_state shows the board, get_decision the current decision, get_card explains a card, get_log shows recent events. Make each decision with choose_option.`

const jsonPrompt = `At each decision you get the game state and the decision as JSON. Answer with a single JSON object and nothing else:
{"reasoning": "<brief justification>", "option_id": <id of the chosen option>}`

func (l *LLM) system() string {
	p := harnessPrompt + "\n\n" + jsonPrompt
	if l.Protocol == ProtocolTools {
		p = harnessPrompt + "\n\n" + toolsPrompt
	}
	if l.Instructions.Text != "" {
		p += "\n\n# Instructions\n\n" + l.Instructions.Text
	}
	return p
}

func (l *LLM) Play(ctx context.Context, s *game.Session) (Usage, error) {
	var u Usage
	msgs := add(s, nil, Message{Role: "system", Content: l.system()}, nil)
	events := []string{"The game is set up."}
	for s.Status() == game.AwaitingDecision {
		msgs = add(s, msgs, Message{Role: "user", Content: l.turnPrompt(s, events)}, nil)
		var err error
		if msgs, events, err = l.decide(ctx, s, msgs, &u); err != nil {
			return u, err
		}
	}
	return u, nil
}

// add appends a message to the conversation and writes it to the trace.
func add(s *game.Session, msgs []Message, m Message, r *Reply) []Message {
	entry := map[string]any{"type": "message", "decision": len(s.Choices) + 1, "message": m}
	if r != nil {
		entry["prompt_tokens"], entry["completion_tokens"], entry["truncated"] = r.InputTokens, r.OutputTokens, r.Truncated
	}
	s.Trace(entry)
	return append(msgs, m)
}

func (l *LLM) turnPrompt(s *game.Session, events []string) string {
	var b strings.Builder
	b.WriteString("Events since your last decision:\n")
	for _, e := range events {
		b.WriteString("- " + e + "\n")
	}
	if l.Protocol == ProtocolJSON {
		b.WriteString("\nState (get_state):\n" + s.Call("get_state", nil) + "\n")
	}
	b.WriteString("\nDecision (get_decision):\n" + s.Call("get_decision", nil))
	return b.String()
}

// decide runs one decision: model requests until choose_option succeeds.
// It returns the extended conversation and the events after the choice.
func (l *LLM) decide(ctx context.Context, s *game.Session, msgs []Message, u *Usage) ([]Message, []string, error) {
	tools := []game.Tool(nil)
	if l.Protocol == ProtocolTools {
		tools = game.Tools
	}
	truncated := 0
	for step := 0; step < l.MaxSteps; step++ {
		reply, err := l.Chat.Complete(ctx, msgs, tools)
		if err != nil {
			return msgs, nil, err
		}
		u.Requests++
		u.InputTokens += reply.InputTokens
		u.OutputTokens += reply.OutputTokens
		if reply.Truncated {
			truncated++
		}
		msgs = add(s, msgs, reply.Message, &reply)

		if l.Protocol == ProtocolJSON {
			args, ok := parseAnswer(reply.Message.Content)
			if !ok {
				msgs = add(s, msgs, Message{Role: "user", Content: "Not a valid answer. " + jsonPrompt}, nil)
				continue
			}
			result := s.Call("choose_option", args)
			if events, ok := chosen(result); ok {
				return msgs, events, nil
			}
			msgs = add(s, msgs, Message{Role: "user", Content: "choose_option failed: " + result}, nil)
			continue
		}

		if len(reply.Message.ToolCalls) == 0 {
			msgs = add(s, msgs, Message{Role: "user", Content: "Make the decision by calling choose_option."}, nil)
			continue
		}
		var events []string
		done := false
		for _, tc := range reply.Message.ToolCalls {
			result := `{"error":"the decision was already made in this reply"}`
			if !done {
				result = s.Call(tc.Function.Name, json.RawMessage(tc.Function.Arguments))
				if tc.Function.Name == "choose_option" {
					events, done = chosen(result)
				}
			}
			msgs = add(s, msgs, Message{Role: "tool", ToolCallID: tc.ID, Content: result}, nil)
		}
		if done {
			return msgs, events, nil
		}
	}
	return msgs, nil, fmt.Errorf("no valid decision after %d model requests (%d cut off at max tokens)", l.MaxSteps, truncated)
}

// chosen reports whether a choose_option result succeeded, with its events.
func chosen(result string) ([]string, bool) {
	var r struct {
		Error  string   `json:"error"`
		Events []string `json:"events"`
	}
	if json.Unmarshal([]byte(result), &r) != nil || r.Error != "" {
		return nil, false
	}
	return r.Events, true
}

var (
	thinkBlock = regexp.MustCompile(`(?s)<think>.*?</think>`)
	jsonObject = regexp.MustCompile(`(?s)\{[^{}]*"option_id"[^{}]*\}`)
	bareID     = regexp.MustCompile(`"?option_id"?\s*[:=]\s*(\d+)`)
)

// parseAnswer extracts {"option_id", "reasoning"} from model text, tolerating
// reasoning blocks, code fences and prose. The last JSON object wins.
func parseAnswer(text string) (json.RawMessage, bool) {
	text = thinkBlock.ReplaceAllString(text, "")
	if objs := jsonObject.FindAllString(text, -1); len(objs) > 0 {
		var a struct {
			OptionID int `json:"option_id"`
		}
		last := objs[len(objs)-1]
		if json.Unmarshal([]byte(last), &a) == nil && a.OptionID > 0 {
			return json.RawMessage(last), true
		}
	}
	if m := bareID.FindAllStringSubmatch(text, -1); len(m) > 0 {
		id, _ := strconv.Atoi(m[len(m)-1][1])
		return json.RawMessage(fmt.Sprintf(`{"option_id":%d}`, id)), id > 0
	}
	return nil, false
}

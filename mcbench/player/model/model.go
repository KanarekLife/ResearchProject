// Package model plays a game as one append-only conversation with a model
// reached through an inference.Client. The model acts only by calling the
// game tools; there is no other protocol.
package model

import (
	"context"
	"encoding/json"
	"fmt"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/integrations/inference"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Model is the LLM player.
//
// A game is one append-only conversation (the agentic loop): a user message
// with the first decision, then the model's replies (with their reasoning)
// and the tool results. Each successful choose_option result carries the
// events and the next decision, so no further user turn repeats it. Nothing is ever removed or rewritten, so each request's
// prompt is a prefix of the next and the provider's prompt cache stays valid.
type Model struct {
	Client       inference.Client
	Label        string // identifies the model in results
	MaxFailures  int    // failed replies allowed per decision
	Instructions instruction.Instructions
}

// maxRequests bounds the replies to one decision, failed or not. Lookups are
// free, so this only guards against a model that never stops looking.
const maxRequests = 100

func (m *Model) Name() string { return "model:" + m.Label }

// Play answers every decision until the game is no longer awaiting one.
func (m *Model) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	var u player.Usage
	g := player.GameFrom(ctx)
	ctx = inference.WithGame(ctx, g.Run, g.Name)
	msgs, err := m.restore(s, player.ResumeFrom(ctx), &u)
	if err != nil {
		return u, fmt.Errorf("%w: resuming: %w", player.ErrInference, err)
	}
	if msgs == nil {
		msgs = m.add(s, nil, inference.Message{Role: inference.RoleSystem, Content: m.system()}, nil)
		msgs = m.add(s, msgs, inference.Message{Role: inference.RoleUser, Content: firstPrompt(s)}, nil)
	}
	for s.Status() == constants.AwaitingDecision {
		if msgs, err = m.decide(ctx, s, msgs, &u); err != nil {
			return u, err
		}
	}
	return u, nil
}

// messageTrace is the trace line of one conversation message. The token
// fields are set only for assistant replies.
type messageTrace struct {
	session.TraceHeader
	Decision     int               `json:"decision"`
	Message      inference.Message `json:"message"`
	InputTokens  int64             `json:"prompt_tokens,omitempty"`
	OutputTokens int64             `json:"completion_tokens,omitempty"`
	CachedTokens int64             `json:"cached_tokens,omitempty"`
	Truncated    bool              `json:"truncated,omitempty"`
}

// add appends a message to the conversation and writes it to the trace.
func (m *Model) add(s *session.Session, msgs []inference.Message, msg inference.Message, r *inference.Reply) []inference.Message {
	entry := &messageTrace{Decision: len(s.Choices) + 1, Message: msg}
	if r != nil {
		entry.InputTokens, entry.OutputTokens = r.InputTokens, r.OutputTokens
		entry.CachedTokens, entry.Truncated = r.CachedTokens, r.Truncated
	}
	s.Trace(constants.TraceMessage, entry)
	return append(msgs, msg)
}

// firstPrompt opens the game with the first decision.
func firstPrompt(s *session.Session) string {
	return "The game is set up.\n\nDecision (" + constants.ToolGetDecision + "):\n" + call(s, constants.ToolGetDecision, nil)
}

// decide runs one decision: model requests until choose_option succeeds.
// A reply fails when it calls no tool or one of its calls returns an error;
// lookups that succeed cost nothing. It returns the extended conversation.
func (m *Model) decide(ctx context.Context, s *session.Session, msgs []inference.Message, u *player.Usage) ([]inference.Message, error) {
	failures, truncated := 0, 0
	for range maxRequests {
		reply, err := m.Client.Complete(ctx, msgs, tools)
		if err != nil {
			return msgs, fmt.Errorf("%w: %w", player.ErrInference, err)
		}
		s.Logger().Debug("model reply", "decision", len(s.Choices)+1, "failures", failures,
			"input_tokens", reply.InputTokens, "output_tokens", reply.OutputTokens, "truncated", reply.Truncated)
		u.Requests++
		u.InputTokens += reply.InputTokens
		u.OutputTokens += reply.OutputTokens
		u.CachedTokens += reply.CachedTokens
		if reply.Truncated {
			truncated++
		}
		msgs = m.add(s, msgs, reply.Message, &reply)

		if len(reply.Message.ToolCalls) == 0 {
			failures++
			content := "Make the decision by calling " + constants.ToolChooseOption + "."
			msgs = m.add(s, msgs, inference.Message{Role: inference.RoleUser, Content: content}, nil)
		} else {
			result := runTools(s, reply.Message.ToolCalls)
			for _, msg := range result.messages {
				msgs = m.add(s, msgs, msg, nil)
			}
			if result.chosen {
				return msgs, nil
			}
			if result.failed {
				failures++
			}
		}
		if failures >= max(1, m.MaxFailures) {
			return msgs, fmt.Errorf("no valid decision after %d failed replies (%d cut off at max tokens)", failures, truncated)
		}
	}
	return msgs, fmt.Errorf("no decision after %d model requests", maxRequests)
}

// toolResult is the outcome of one assistant reply's tool calls.
type toolResult struct {
	messages []inference.Message // tool result messages to append
	chosen   bool                // a choose_option succeeded
	failed   bool                // a call returned an error
}

// runTools executes the tool calls in order. Only the first choose_option is
// applied; later ones get an error result.
func runTools(s *session.Session, calls []inference.ToolCall) toolResult {
	var res toolResult
	for _, tc := range calls {
		out := `{"error":"the decision was already made in this reply"}`
		if !res.chosen {
			var failed bool
			out, res.chosen, failed = exec(s, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			res.failed = res.failed || failed
		}
		res.messages = append(res.messages, inference.Message{Role: inference.RoleTool, ToolCallID: tc.ID, Content: out})
	}
	return res
}

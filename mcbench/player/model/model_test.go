package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/integrations/inference"
)

// scripted is an inference.Client that answers from a function and records
// every conversation it was sent.
type scripted struct {
	reply func(call int, msgs []inference.Message) inference.Message
	seen  [][]inference.Message
}

func (c *scripted) Complete(_ context.Context, msgs []inference.Message, _ []inference.Tool) (inference.Reply, error) {
	c.seen = append(c.seen, append([]inference.Message(nil), msgs...))
	return inference.Reply{Message: c.reply(len(c.seen), msgs), InputTokens: 10, OutputTokens: 2}, nil
}

func toolCall(name, args string) inference.ToolCall {
	tc := inference.ToolCall{ID: "id", Type: "function"}
	tc.Function.Name, tc.Function.Arguments = name, args
	return tc
}

func assistant(calls ...inference.ToolCall) inference.Message {
	return inference.Message{Role: inference.RoleAssistant, ToolCalls: calls}
}

func newSession(t *testing.T) *session.Session {
	t.Helper()
	scs, err := scenario.LoadDir("../../data/scenarios", "../../data")
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.New(scs[0], scs[0].Seeds[0], session.Options{MaxRounds: scs[0].MaxRounds, MaxDecisions: scs[0].MaxDecisions})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func choose(id int) inference.ToolCall {
	return toolCall(constants.ToolChooseOption, fmt.Sprintf(`{"option_id": %d}`, id))
}

func TestPlayRunsTheGameToTheEnd(t *testing.T) {
	c := &scripted{reply: func(int, []inference.Message) inference.Message { return assistant(choose(1)) }}
	s := newSession(t)
	u, err := (&Model{Client: c, Label: "x", MaxSteps: 3}).Play(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Status() == constants.AwaitingDecision {
		t.Fatalf("game still awaiting a decision")
	}
	if u.Requests != len(s.Choices) || u.InputTokens != int64(10*u.Requests) {
		t.Errorf("usage %+v for %d choices", u, len(s.Choices))
	}
}

// The first decision is the only user turn: later decisions arrive in the
// choose_option results and are not repeated.
func TestDecisionIsSentOnce(t *testing.T) {
	c := &scripted{reply: func(int, []inference.Message) inference.Message { return assistant(choose(1)) }}
	(&Model{Client: c, MaxSteps: 3}).Play(context.Background(), newSession(t))
	users := 0
	for _, msg := range c.seen[len(c.seen)-1] {
		if msg.Role == inference.RoleUser {
			users++
		}
	}
	if users != 1 {
		t.Errorf("%d user messages, want 1", users)
	}
}

// The conversation is append-only: each request's messages extend the
// previous request's, which keeps the provider's prompt cache valid.
func TestConversationIsAppendOnly(t *testing.T) {
	c := &scripted{reply: func(int, []inference.Message) inference.Message { return assistant(choose(1)) }}
	(&Model{Client: c, MaxSteps: 3}).Play(context.Background(), newSession(t))
	for i := 1; i < len(c.seen) && i < 30; i++ {
		prev, cur := c.seen[i-1], c.seen[i]
		if len(cur) <= len(prev) {
			t.Fatalf("request %d has %d messages after %d", i+1, len(cur), len(prev))
		}
		for j := range prev {
			if prev[j].Content != cur[j].Content || prev[j].Role != cur[j].Role {
				t.Fatalf("request %d rewrote message %d", i+1, j)
			}
		}
	}
}

func TestInvalidChoiceIsReportedAndRetried(t *testing.T) {
	c := &scripted{reply: func(call int, _ []inference.Message) inference.Message {
		if call == 1 {
			return assistant(choose(999))
		}
		return assistant(choose(1))
	}}
	s := newSession(t)
	if _, err := (&Model{Client: c, MaxSteps: 3}).Play(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if s.Invalid != 1 {
		t.Errorf("invalid = %d, want 1", s.Invalid)
	}
	last := c.seen[1][len(c.seen[1])-1]
	if last.Role != inference.RoleTool || !strings.Contains(last.Content, "does not exist") {
		t.Errorf("model was not told why its choice failed: %+v", last)
	}
}

func TestReplyWithoutToolCallIsPrompted(t *testing.T) {
	c := &scripted{reply: func(int, []inference.Message) inference.Message {
		return inference.Message{Role: inference.RoleAssistant, Content: "thinking"}
	}}
	_, err := (&Model{Client: c, MaxSteps: 2}).Play(context.Background(), newSession(t))
	if err == nil || !strings.Contains(err.Error(), "no valid decision after 2") {
		t.Fatalf("err = %v, want the step-limit error", err)
	}
	nudge := c.seen[1][len(c.seen[1])-1]
	if nudge.Role != inference.RoleUser || !strings.Contains(nudge.Content, constants.ToolChooseOption) {
		t.Errorf("no reminder to call %s: %+v", constants.ToolChooseOption, nudge)
	}
}

// Only the first choose_option of a reply is applied.
func TestOnlyFirstChooseIsApplied(t *testing.T) {
	s := newSession(t)
	res := runTools(s, []inference.ToolCall{choose(1), choose(1)})
	if !res.chosen || len(s.Choices) != 1 || len(res.messages) != 2 {
		t.Fatalf("chosen=%v choices=%d messages=%d", res.chosen, len(s.Choices), len(res.messages))
	}
	var r map[string]any
	json.Unmarshal([]byte(res.messages[1].Content), &r)
	if _, ok := r["error"]; !ok {
		t.Errorf("second choose_option was not refused: %v", r)
	}
}

func TestSystemPromptAppendsInstructions(t *testing.T) {
	m := &Model{}
	if m.system() != systemPrompt {
		t.Error("no instructions: the prompt must be the bare harness prompt")
	}
	m.Instructions.Text = "RULES"
	if got := m.system(); !strings.HasPrefix(got, systemPrompt) || !strings.HasSuffix(got, "RULES") {
		t.Errorf("system prompt = %q", got)
	}
}

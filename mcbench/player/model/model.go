// Package model plays a game as one append-only conversation with a model
// reached through an inference.Client. The model acts only by calling the
// game tools; there is no other protocol.
package model

import (
	"context"
	"fmt"
	"strings"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/integrations/inference"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Model is the LLM player.
//
// A game is one append-only conversation (the agentic loop): each decision
// adds a user message with the events since the last decision and the
// current decision, then the model's replies (with their reasoning) and any
// tool results. Nothing is ever removed or rewritten, so each request's
// prompt is a prefix of the next and the provider's prompt cache stays valid.
type Model struct {
	Client       inference.Client
	Label        string // identifies the model in results
	MaxSteps     int    // model requests allowed per decision
	Instructions instruction.Instructions
}

func (m *Model) Name() string { return "model:" + m.Label }

// Play answers every decision until the game is no longer awaiting one.
func (m *Model) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	var u player.Usage
	msgs := m.add(s, nil, inference.Message{Role: inference.RoleSystem, Content: m.system()}, nil)
	events := []string{"The game is set up."}
	for s.Status() == session.AwaitingDecision {
		msgs = m.add(s, msgs, inference.Message{Role: inference.RoleUser, Content: m.turnPrompt(s, events)}, nil)
		var err error
		if msgs, events, err = m.decide(ctx, s, msgs, &u); err != nil {
			return u, err
		}
	}
	return u, nil
}

// add appends a message to the conversation and writes it to the trace.
func (m *Model) add(s *session.Session, msgs []inference.Message, msg inference.Message, r *inference.Reply) []inference.Message {
	entry := map[string]any{
		constants.TraceKeyType:     constants.TraceMessage,
		constants.TraceKeyDecision: len(s.Choices) + 1,
		constants.TraceKeyMessage:  msg,
	}
	if r != nil {
		entry[constants.TraceKeyInputTok], entry[constants.TraceKeyOutputTok] = r.InputTokens, r.OutputTokens
		entry[constants.TraceKeyCachedTok], entry[constants.TraceKeyTruncated] = r.CachedTokens, r.Truncated
	}
	s.Trace(entry)
	return append(msgs, msg)
}

func (m *Model) turnPrompt(s *session.Session, events []string) string {
	var b strings.Builder
	b.WriteString("Events since your last decision:\n")
	for _, e := range events {
		b.WriteString("- " + e + "\n")
	}
	b.WriteString("\nDecision (" + constants.ToolGetDecision + "):\n" + s.Call(constants.ToolGetDecision, nil))
	return b.String()
}

// decide runs one decision: model requests until choose_option succeeds.
// It returns the extended conversation and the events after the choice.
func (m *Model) decide(ctx context.Context, s *session.Session, msgs []inference.Message, u *player.Usage) ([]inference.Message, []string, error) {
	truncated := 0
	for step := 0; step < m.MaxSteps; step++ {
		reply, err := m.Client.Complete(ctx, msgs, toolSchemas())
		if err != nil {
			return msgs, nil, err
		}
		s.Logger().Debug("model reply", "decision", len(s.Choices)+1, "step", step+1,
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
			content := "Make the decision by calling " + constants.ToolChooseOption + "."
			msgs = m.add(s, msgs, inference.Message{Role: inference.RoleUser, Content: content}, nil)
			continue
		}
		result := runTools(s, reply.Message.ToolCalls)
		for _, msg := range result.messages {
			msgs = m.add(s, msgs, msg, nil)
		}
		if result.chosen {
			return msgs, result.events, nil
		}
	}
	return msgs, nil, fmt.Errorf("no valid decision after %d model requests (%d cut off at max tokens)", m.MaxSteps, truncated)
}

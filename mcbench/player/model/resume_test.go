package model

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"mcbench/constants"
	"mcbench/data"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/integrations/inference"
	"mcbench/player"
)

func newTracedSession(t *testing.T, trace *bytes.Buffer) *session.Session {
	t.Helper()
	scs, err := scenario.LoadAll(data.FS)
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.New(scs[0], scs[0].Seeds[0], session.Options{MaxRounds: scs[0].MaxRounds, MaxDecisions: scs[0].MaxDecisions, Trace: trace})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func keys(s *session.Session) []string {
	var out []string
	for _, c := range s.Choices {
		out = append(out, c.Key)
	}
	return out
}

// A game cut short and resumed from its trace must reach the game an
// uninterrupted player reaches, and the model must be sent the conversation it
// had before the cut.
func TestResumeContinuesACutGame(t *testing.T) {
	first := func(int, []inference.Message) inference.Message { return assistant(choose(1)) }
	whole := newSession(t)
	if _, err := (&Model{Client: &scripted{reply: first}, MaxFailures: 3}).Play(context.Background(), whole); err != nil {
		t.Fatal(err)
	}

	var cut bytes.Buffer
	dying := &failAfter{scripted: scripted{reply: first}, n: 15}
	_, err := (&Model{Client: dying, MaxFailures: 3}).Play(context.Background(), newTracedSession(t, &cut))
	if !errors.Is(err, player.ErrInference) {
		t.Fatalf("cut game: %v", err)
	}
	cut.Write([]byte(`{"type":"message","decision":9`)) // the line being written when the run was killed

	var resumed bytes.Buffer
	c := &scripted{reply: first}
	s := newTracedSession(t, &resumed)
	if _, err := (&Model{Client: c, MaxFailures: 3}).Play(player.WithResume(context.Background(), cut.Bytes()), s); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys(s), keys(whole)) {
		t.Errorf("resumed game made different choices than the uninterrupted one")
	}
	if s.Status() == constants.AwaitingDecision {
		t.Errorf("resumed game did not finish")
	}
	if got := len(c.seen[0]); got < 5 || c.seen[0][0].Role != inference.RoleSystem {
		t.Errorf("first request after resuming carried %d messages", got)
	}
}

// failAfter fails every request after its first n.
type failAfter struct {
	scripted
	n int
}

func (f *failAfter) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	if len(f.seen) >= f.n {
		return inference.Reply{}, errors.New("cut")
	}
	return f.scripted.Complete(ctx, msgs, tools)
}

func TestCompletePrefixDropsAnUnansweredCall(t *testing.T) {
	user := traceLine{Message: inference.Message{Role: inference.RoleUser}}
	call := traceLine{Message: assistant(choose(1))}
	result := traceLine{Message: inference.Message{Role: inference.RoleTool, ToolCallID: "id"}}
	if n := completePrefix([]traceLine{user, call, result, call}); n != 3 {
		t.Errorf("prefix %d, want 3", n)
	}
}

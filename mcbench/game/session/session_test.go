package session_test

import (
	"strings"
	"testing"

	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/game/session"
)

func newSession(t *testing.T, opts session.Options) *session.Session {
	t.Helper()
	scs, err := scenario.LoadDir("../../data/scenarios", "../../data")
	if err != nil {
		t.Fatal(err)
	}
	sc := scs[0]
	opts.MaxRounds, opts.MaxDecisions = sc.MaxRounds, sc.MaxDecisions
	s, err := session.New(sc, sc.Seeds[0], opts)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func keys(d *session.Decision) []string {
	var out []string
	for _, o := range d.Options {
		out = append(out, o.Key)
	}
	return out
}

func TestChooseRejectsUnknownOptions(t *testing.T) {
	s := newSession(t, session.Options{})
	n := len(s.Decision().Options)
	for _, id := range []int{-1, 0, n + 1} {
		if _, err := s.Choose(id, ""); err == nil {
			t.Errorf("option %d accepted, want an error", id)
		}
	}
	if s.Invalid != 3 || len(s.Choices) != 0 {
		t.Errorf("invalid=%d choices=%d, want 3 and 0", s.Invalid, len(s.Choices))
	}
	if s.Status() != constants.AwaitingDecision {
		t.Errorf("status %s after rejected choices, want %s", s.Status(), constants.AwaitingDecision)
	}
}

func TestDecisionLimitEndsTheGame(t *testing.T) {
	s := limited(t, 2)
	for i := 0; i < 2; i++ {
		if _, err := s.Choose(1, ""); err != nil {
			t.Fatal(err)
		}
	}
	if s.Status() != constants.DecisionLimit || s.Decision() != nil {
		t.Fatalf("status %s, decision %v; want %s and no decision", s.Status(), s.Decision(), constants.DecisionLimit)
	}
	if _, err := s.Choose(1, ""); err == nil {
		t.Error("choose after the game ended succeeded")
	}
}

func limited(t *testing.T, n int) *session.Session {
	t.Helper()
	scs, err := scenario.LoadDir("../../data/scenarios", "../../data")
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.New(scs[0], scs[0].Seeds[0], session.Options{MaxDecisions: n})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// The option order is a function of seed and sample only, and the sample
// actually changes it.
func TestOptionOrder(t *testing.T) {
	order := func(sample uint64) string {
		s := newSession(t, session.Options{ShuffleOptions: true, Sample: sample})
		s.Choose(1, "") // mulligan; the turn decision has many options
		return strings.Join(keys(s.Decision()), "|")
	}
	if order(0) != order(0) {
		t.Fatal("same sample gave two different option orders")
	}
	distinct := map[string]bool{}
	for sample := uint64(0); sample < 6; sample++ {
		distinct[order(sample)] = true
	}
	if len(distinct) < 2 {
		t.Error("six samples all gave the same option order")
	}
}

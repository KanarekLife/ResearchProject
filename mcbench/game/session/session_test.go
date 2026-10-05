package session_test

import (
	"encoding/json"
	"slices"
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
	if s.Status() != session.AwaitingDecision {
		t.Errorf("status %s after rejected choices, want %s", s.Status(), session.AwaitingDecision)
	}
}

func TestDecisionLimitEndsTheGame(t *testing.T) {
	s := limited(t, 2)
	for i := 0; i < 2; i++ {
		if _, err := s.Choose(1, ""); err != nil {
			t.Fatal(err)
		}
	}
	if s.Status() != session.DecisionLimit || s.Decision() != nil {
		t.Fatalf("status %s, decision %v; want %s and no decision", s.Status(), s.Decision(), session.DecisionLimit)
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

func call(t *testing.T, s *session.Session, tool, args string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(s.Call(tool, json.RawMessage(args))), &out); err != nil {
		t.Fatalf("%s: result is not a JSON object: %v", tool, err)
	}
	return out
}

func TestCallReportsErrorsAsJSON(t *testing.T) {
	s := newSession(t, session.Options{})
	for name, tc := range map[string]struct{ tool, args string }{
		"unknown tool":        {"no_such_tool", ""},
		"malformed arguments": {constants.ToolChooseOption, `{"option_id": "x"}`},
		"option out of range": {constants.ToolChooseOption, `{"option_id": 999}`},
		"unknown card":        {constants.ToolGetCard, `{"name": "No Such Card"}`},
	} {
		if _, ok := call(t, s, tc.tool, tc.args)["error"]; !ok {
			t.Errorf("%s: no error in result", name)
		}
	}
}

func TestCallChooseAdvancesTheGame(t *testing.T) {
	s := newSession(t, session.Options{})
	r := call(t, s, constants.ToolChooseOption, `{"option_id": 1, "reasoning": "test"}`)
	if _, ok := r["events"]; !ok {
		t.Fatalf("choose_option result has no events: %v", r)
	}
	if len(s.Choices) != 1 || s.Choices[0].Reasoning != "test" {
		t.Errorf("choices = %+v, want one choice carrying the reasoning", s.Choices)
	}
	log := call(t, s, constants.ToolGetLog, `{"last": 2}`)["log"].([]any)
	if len(log) > 2 {
		t.Errorf("get_log last=2 returned %d lines", len(log))
	}
}

func TestContractHasEveryTool(t *testing.T) {
	names := []string{}
	for _, tool := range session.Tools {
		names = append(names, tool.Name)
	}
	for _, want := range []string{constants.ToolGetState, constants.ToolGetDecision, constants.ToolChooseOption, constants.ToolGetLog, constants.ToolGetCard} {
		if !slices.Contains(names, want) {
			t.Errorf("contract lacks tool %s", want)
		}
	}
}

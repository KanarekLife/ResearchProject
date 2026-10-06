package model

import (
	"encoding/json"
	"strings"
	"testing"

	"mcbench/constants"
	"mcbench/game/session"
)

func callJSON(t *testing.T, s *session.Session, tool, args string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(call(s, tool, json.RawMessage(args))), &out); err != nil {
		t.Fatalf("%s: result is not a JSON object: %v", tool, err)
	}
	return out
}

func TestCallReportsErrorsAsJSON(t *testing.T) {
	s := newSession(t)
	for name, tc := range map[string]struct{ tool, args string }{
		"unknown tool":        {"no_such_tool", ""},
		"malformed arguments": {constants.ToolChooseOption, `{"option_id": "x"}`},
		"option out of range": {constants.ToolChooseOption, `{"option_id": 999}`},
		"unknown card":        {constants.ToolGetCard, `{"name": "No Such Card"}`},
	} {
		if _, ok := callJSON(t, s, tc.tool, tc.args)["error"]; !ok {
			t.Errorf("%s: no error in result", name)
		}
	}
}

func TestCallChooseAdvancesTheGame(t *testing.T) {
	s := newSession(t)
	r := callJSON(t, s, constants.ToolChooseOption, `{"option_id": 1, "reasoning": "test"}`)
	if _, ok := r["events"]; !ok {
		t.Fatalf("choose_option result has no events: %v", r)
	}
	if len(s.Choices) != 1 || s.Choices[0].Reasoning != "test" {
		t.Errorf("choices = %+v, want one choice carrying the reasoning", s.Choices)
	}
	log := callJSON(t, s, constants.ToolGetLog, `{"last": 2}`)["log"].([]any)
	if len(log) > 2 {
		t.Errorf("get_log last=2 returned %d lines", len(log))
	}
}

// Every advertised tool must be executable: a schema without a call case
// would leave the model with an "unknown tool" error at runtime.
func TestEveryToolExecutes(t *testing.T) {
	s := newSession(t)
	for _, tool := range tools {
		if out := call(s, tool.Name, nil); strings.Contains(out, "unknown tool") {
			t.Errorf("%s is advertised but call does not handle it", tool.Name)
		}
	}
}

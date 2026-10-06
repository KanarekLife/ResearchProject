package model

import (
	"encoding/json"
	"fmt"

	"mcbench/constants"
	"mcbench/game/session"
)

// Tool-result keys.
const (
	resultStatus   = "status"
	resultDecision = "decision"
	resultEvents   = "events"
	resultError    = "error"
	resultLog      = "log"
	// defaultLogLines is how much of the log get_log returns by default.
	defaultLogLines = 20
)

// call runs a tool with JSON arguments against the game and returns its JSON
// result. Errors are returned as {"error": "..."} so a model can read and
// correct them.
func call(s *session.Session, name string, args json.RawMessage) string {
	out, _ := exec(s, name, args)
	return out
}

// exec is call that also reports whether a choose_option succeeded.
func exec(s *session.Session, name string, args json.RawMessage) (out string, chosen bool) {
	var a struct {
		OptionID  int    `json:"option_id"`
		Reasoning string `json:"reasoning"`
		Last      int    `json:"last"`
		Name      string `json:"name"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &a); err != nil {
			return errJSON(fmt.Errorf("invalid arguments: %v", err)), false
		}
	}
	switch name {
	case constants.ToolGetState:
		return toJSON(s.View()), false
	case constants.ToolGetDecision:
		return toJSON(decisionResult(s)), false
	case constants.ToolChooseOption:
		events, err := s.Choose(a.OptionID, a.Reasoning)
		if err != nil {
			return errJSON(err), false
		}
		r := decisionResult(s)
		r[resultEvents] = events
		return toJSON(r), true
	case constants.ToolGetLog:
		if a.Last <= 0 {
			a.Last = defaultLogLines
		}
		return toJSON(map[string]any{resultLog: s.Log(a.Last)}), false
	case constants.ToolGetCard:
		ref, err := session.Reference(a.Name)
		if err != nil {
			return errJSON(err), false
		}
		return toJSON(ref), false
	}
	return errJSON(fmt.Errorf("unknown tool %q", name)), false
}

func decisionResult(s *session.Session) map[string]any {
	r := map[string]any{resultStatus: s.Status()}
	if d := s.Decision(); d != nil {
		r[resultDecision] = d
	}
	return r
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func errJSON(err error) string { return toJSON(map[string]string{resultError: err.Error()}) }

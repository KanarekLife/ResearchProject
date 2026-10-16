package model

import (
	"encoding/json"
	"fmt"
	"strings"

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
	out, _, _ := exec(s, name, args)
	return out
}

// exec is call that also reports whether a choose_option succeeded and
// whether the call failed (its result is an error).
func exec(s *session.Session, name string, args json.RawMessage) (out string, chosen, failed bool) {
	var a struct {
		OptionID  int    `json:"option_id"`
		Reasoning string `json:"reasoning"`
		Last      int    `json:"last"`
		Name      string `json:"name"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &a); err != nil {
			return errJSON(fmt.Errorf("invalid arguments: %v", err)), false, true
		}
	}
	switch name {
	case constants.ToolGetState:
		return toJSON(s.View()), false, false
	case constants.ToolGetDecision:
		return toJSON(decisionResult(s)), false, false
	case constants.ToolChooseOption:
		events, err := s.Choose(a.OptionID, a.Reasoning)
		if err != nil {
			return errJSON(err), false, true
		}
		r := decisionResult(s)
		r[resultEvents] = withoutChoiceEcho(events)
		return toJSON(r), true, false
	case constants.ToolGetLog:
		if a.Last <= 0 {
			a.Last = defaultLogLines
		}
		return toJSON(map[string]any{resultLog: s.Log(a.Last)}), false, false
	case constants.ToolGetCard:
		ref, err := session.Reference(a.Name)
		if err != nil {
			return errJSON(err), false, true
		}
		return toJSON(ref), false, false
	}
	return errJSON(fmt.Errorf("unknown tool %q", name)), false, true
}

func decisionResult(s *session.Session) map[string]any {
	r := map[string]any{resultStatus: s.Status()}
	if d := s.Decision(); d != nil {
		r[resultDecision] = compactDecision(d)
	}
	return r
}

// decisionView is a decision as the model reads it. Options are one string
// each ("2) Defend with Spider-Man"), without the engine key: the result stays
// in the conversation for the rest of the game, so every byte is resent on
// every later request.
type decisionView struct {
	Kind    string   `json:"kind"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

func compactDecision(d *session.Decision) decisionView {
	v := decisionView{Kind: d.Kind, Prompt: d.Prompt, Options: make([]string, len(d.Options))}
	for i, o := range d.Options {
		v.Options[i] = fmt.Sprintf("%d) %s", o.ID, o.Text)
	}
	return v
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func errJSON(err error) string { return toJSON(map[string]string{resultError: err.Error()}) }

// withoutChoiceEcho drops the "Player chose" line: the model just made that
// choice, so it only repeats it. An auto-resolved choice keeps its line.
func withoutChoiceEcho(events []string) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		if !strings.HasPrefix(e, constants.EventChose) {
			out = append(out, e)
		}
	}
	return out
}

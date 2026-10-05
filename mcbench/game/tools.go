package game

import (
	"encoding/json"
	"fmt"

	"mcbench/cards"
	"mcbench/engine"
)

// Tool describes one tool of the contract. Parameters is a JSON Schema.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

func object(props map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
}

// Tools is the agent <-> game contract. See docs/architecture.md.
var Tools = []Tool{
	{
		Name:        "get_state",
		Description: "Get the public game state: your hero, hand, cards in play, the villain, schemes, minions and pile sizes.",
		Parameters:  object(map[string]any{}),
	},
	{
		Name:        "get_decision",
		Description: "Get the decision you must make now and its numbered legal options, or the game status when it is over.",
		Parameters:  object(map[string]any{}),
	},
	{
		Name:        "choose_option",
		Description: "Make the current decision. Returns the events that followed and the next decision (or the final status).",
		Parameters: object(map[string]any{
			"option_id": map[string]any{"type": "integer", "description": "id of a listed option"},
			"reasoning": map[string]any{"type": "string", "description": "brief justification (recorded, not used by the game)"},
		}, "option_id"),
	},
	{
		Name:        "get_log",
		Description: "Get the last lines of the game log.",
		Parameters: object(map[string]any{
			"last": map[string]any{"type": "integer", "description": "number of lines (default 20)"},
		}),
	},
	{
		Name:        "get_card",
		Description: "Get the reference text and stats of a card by name.",
		Parameters:  object(map[string]any{"name": map[string]any{"type": "string"}}, "name"),
	},
}

// Call runs a tool with JSON arguments and returns its JSON result. Errors
// are returned as {"error": "..."} so a model can read and correct them.
func (s *Session) Call(name string, args json.RawMessage) string {
	var a struct {
		OptionID  int    `json:"option_id"`
		Reasoning string `json:"reasoning"`
		Last      int    `json:"last"`
		Name      string `json:"name"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &a); err != nil {
			return errJSON(fmt.Errorf("invalid arguments: %v", err))
		}
	}
	switch name {
	case "get_state":
		return toJSON(s.View())
	case "get_decision":
		return toJSON(s.decisionResult())
	case "choose_option":
		events, err := s.Choose(a.OptionID, a.Reasoning)
		if err != nil {
			return errJSON(err)
		}
		r := s.decisionResult()
		r["events"] = events
		return toJSON(r)
	case "get_log":
		if a.Last <= 0 {
			a.Last = 20
		}
		return toJSON(map[string]any{"log": s.Log(a.Last)})
	case "get_card":
		d, err := cards.Get(a.Name)
		if err != nil {
			return errJSON(err)
		}
		if d.IsPlayerCard() {
			return toJSON(handCard(&engine.Card{Def: d}))
		}
		return toJSON(map[string]any{"name": d.Name, "type": d.Type, "hp": d.HP, "atk": d.ATK, "sch": d.SCH, "boost": d.Boost, "text": d.Text})
	}
	return errJSON(fmt.Errorf("unknown tool %q", name))
}

func (s *Session) decisionResult() map[string]any {
	r := map[string]any{"status": s.Status()}
	if d := s.Decision(); d != nil {
		r["decision"] = d
	}
	return r
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func errJSON(err error) string { return toJSON(map[string]string{"error": err.Error()}) }

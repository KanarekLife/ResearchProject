package session

import (
	"encoding/json"
	"fmt"

	"mcbench/constants"
	"mcbench/game/cards"
	"mcbench/game/engine"
)

// JSON Schema keywords and tool-result keys.
const (
	schemaObject      = "object"
	schemaType        = "type"
	schemaProperties  = "properties"
	schemaRequired    = "required"
	schemaAdditional  = "additionalProperties"
	schemaDescription = "description"
	schemaInteger     = "integer"
	schemaString      = "string"
	resultStatus      = "status"
	resultDecision    = "decision"
	resultEvents      = "events"
	resultError       = "error"
	resultLog         = "log"
	// defaultLogLines is how much of the log get_log returns by default.
	defaultLogLines = 20
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
	return map[string]any{
		schemaType:       schemaObject,
		schemaProperties: props,
		schemaRequired:   required,
		schemaAdditional: false,
	}
}

// Tools is the player <-> game contract. See ../../../docs/architecture.md.
var Tools = []Tool{
	{
		Name:        constants.ToolGetState,
		Description: "Get the public game state: your hero, hand, cards in play, the villain, schemes, minions and pile sizes.",
		Parameters:  object(map[string]any{}),
	},
	{
		Name:        constants.ToolGetDecision,
		Description: "Get the decision you must make now and its numbered legal options, or the game status when it is over.",
		Parameters:  object(map[string]any{}),
	},
	{
		Name:        constants.ToolChooseOption,
		Description: "Make the current decision. Returns the events that followed and the next decision (or the final status).",
		Parameters: object(map[string]any{
			constants.ArgOptionID:  map[string]any{schemaType: schemaInteger, schemaDescription: "id of a listed option"},
			constants.ArgReasoning: map[string]any{schemaType: schemaString, schemaDescription: "brief justification (recorded, not used by the game)"},
		}, constants.ArgOptionID),
	},
	{
		Name:        constants.ToolGetLog,
		Description: "Get the last lines of the game log.",
		Parameters: object(map[string]any{
			constants.ArgLast: map[string]any{schemaType: schemaInteger, schemaDescription: "number of lines (default 20)"},
		}),
	},
	{
		Name:        constants.ToolGetCard,
		Description: "Get the reference text and stats of a card by name.",
		Parameters:  object(map[string]any{constants.ArgName: map[string]any{schemaType: schemaString}}, constants.ArgName),
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
	case constants.ToolGetState:
		return toJSON(s.View())
	case constants.ToolGetDecision:
		return toJSON(s.decisionResult())
	case constants.ToolChooseOption:
		return s.chooseResult(a.OptionID, a.Reasoning)
	case constants.ToolGetLog:
		if a.Last <= 0 {
			a.Last = defaultLogLines
		}
		return toJSON(map[string]any{resultLog: s.Log(a.Last)})
	case constants.ToolGetCard:
		return cardResult(a.Name)
	}
	return errJSON(fmt.Errorf("unknown tool %q", name))
}

// chooseResult answers via choose_option, adding the events to the reply.
func (s *Session) chooseResult(optionID int, reasoning string) string {
	events, err := s.Choose(optionID, reasoning)
	if err != nil {
		return errJSON(err)
	}
	r := s.decisionResult()
	r[resultEvents] = events
	return toJSON(r)
}

func cardResult(name string) string {
	d, err := cards.Get(name)
	if err != nil {
		return errJSON(err)
	}
	if d.IsPlayerCard() {
		return toJSON(handCard(&engine.Card{Def: d}))
	}
	return toJSON(map[string]any{"name": d.Name, "type": d.Type, "hp": d.HP, "atk": d.ATK, "sch": d.SCH, "boost": d.Boost, "text": d.Text})
}

func (s *Session) decisionResult() map[string]any {
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

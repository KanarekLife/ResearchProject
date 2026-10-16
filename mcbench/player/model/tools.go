package model

import (
	"mcbench/constants"
	"mcbench/integrations/inference"
)

// JSON Schema keywords.
const (
	schemaObject      = "object"
	schemaType        = "type"
	schemaProperties  = "properties"
	schemaRequired    = "required"
	schemaAdditional  = "additionalProperties"
	schemaDescription = "description"
	schemaInteger     = "integer"
	schemaString      = "string"
)

// tools is the player <-> game contract: the five tools the model may call,
// executed against a session by call. See ../../../docs/architecture.md.
var tools = []inference.Tool{
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
		Description: "Make the current decision. Call it once per reply: a second call in the same reply is rejected, because the options change after each choice. Returns the events that followed and the next decision (or the final status).",
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
		Description: "Get the reference text and stats of a card by name or code. Close spellings and partial names are accepted; when several cards match, it lists them with their codes so you can ask for one.",
		Parameters:  object(map[string]any{constants.ArgName: map[string]any{schemaType: schemaString}}, constants.ArgName),
	},
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

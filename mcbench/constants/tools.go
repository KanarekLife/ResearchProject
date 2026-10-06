package constants

// Tool names of the player <-> game contract. These are the function names a
// model calls; they also identify the tools in records and traces.
const (
	ToolGetState     = "get_state"
	ToolGetDecision  = "get_decision"
	ToolChooseOption = "choose_option"
	ToolGetLog       = "get_log"
	ToolGetCard      = "get_card"
)

// Argument keys of the tool schema.
const (
	ArgOptionID  = "option_id"
	ArgReasoning = "reasoning"
	ArgLast      = "last"
	ArgName      = "name"
)

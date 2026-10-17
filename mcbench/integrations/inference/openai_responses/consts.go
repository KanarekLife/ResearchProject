package openairesponses

// DefaultBaseURL is OpenAI's own endpoint.
const DefaultBaseURL = "https://api.openai.com/v1"

// Wire-format names of the OpenAI Responses API.
const (
	pathResponses = "/responses"

	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "

	fieldModel           = "model"
	fieldInstructions    = "instructions"
	fieldInput           = "input"
	fieldMaxOutputTokens = "max_output_tokens"
	fieldTemperature     = "temperature"
	fieldReasoning       = "reasoning"
	fieldEffort          = "effort"
	fieldTools           = "tools"
	fieldToolChoice      = "tool_choice"
	fieldStore           = "store"
	fieldType            = "type"
	fieldName            = "name"
	fieldDescription     = "description"
	fieldParameters      = "parameters"
	toolTypeFunction     = "function"
	toolChoiceAuto       = "auto"

	roleUser      = "user"
	roleAssistant = "assistant"

	itemMessage            = "message"
	itemFunctionCall       = "function_call"
	itemFunctionCallOutput = "function_call_output"
	itemReasoning          = "reasoning"
	contentOutputText      = "output_text"

	statusIncomplete      = "incomplete"
	reasonMaxOutputTokens = "max_output_tokens"
)

package openairesponses

// DefaultBaseURL is OpenAI's own endpoint.
const DefaultBaseURL = "https://api.openai.com/v1"

// Wire-format names of the OpenAI Responses API.
const (
	pathResponses = "/responses"

	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "

	fieldStream          = "stream"
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
	fieldInclude         = "include"
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

	includeEncryptedReasoning = "reasoning.encrypted_content"

	eventCompleted  = "response.completed"
	eventIncomplete = "response.incomplete"
	eventFailed     = "response.failed"
	eventError      = "error"

	statusIncomplete      = "incomplete"
	reasonMaxOutputTokens = "max_output_tokens"
)

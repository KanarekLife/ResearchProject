package openaicompatible

// Wire-format names of the OpenAI-compatible chat-completions API.
const (
	pathChatCompletions = "/chat/completions"

	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	headerRetryAfter    = "Retry-After"
	mimeJSON            = "application/json"
	bearerPrefix        = "Bearer "

	fieldModel               = "model"
	fieldMessages            = "messages"
	fieldMaxTokens           = "max_tokens"
	fieldMaxCompletionTokens = "max_completion_tokens"
	fieldTemperature         = "temperature"
	fieldTools               = "tools"
	fieldToolChoice          = "tool_choice"
	fieldType                = "type"
	fieldFunction            = "function"
	toolChoiceAuto           = "auto"

	finishReasonLength = "length"
)

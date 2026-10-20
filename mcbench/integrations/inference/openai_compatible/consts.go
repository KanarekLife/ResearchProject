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
	fieldReasoningEffort     = "reasoning_effort"
	fieldReasoningBudget     = "reasoning_budget_tokens"
	fieldThinkingBudget      = "thinking_budget_tokens"
	fieldStream              = "stream"
	fieldStreamOptions       = "stream_options"
	fieldIncludeUsage        = "include_usage"
	fieldTools               = "tools"
	fieldToolChoice          = "tool_choice"
	fieldType                = "type"
	fieldFunction            = "function"
	toolChoiceAuto           = "auto"

	finishReasonLength = "length"
)

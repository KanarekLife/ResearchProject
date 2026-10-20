package anthropic

// DefaultBaseURL is Anthropic's own endpoint.
const DefaultBaseURL = "https://api.anthropic.com/v1"

// Wire-format names of the Messages API.
const (
	apiVersion = "2023-06-01"

	pathMessages = "/messages"

	headerAPIKey  = "x-api-key"
	headerVersion = "anthropic-version"

	roleUser      = "user"
	roleAssistant = "assistant"

	blockText       = "text"
	blockToolUse    = "tool_use"
	blockToolResult = "tool_result"
	blockThinking   = "thinking"

	fieldStream      = "stream"
	fieldModel       = "model"
	fieldMaxTokens   = "max_tokens"
	fieldTemperature = "temperature"
	fieldSystem      = "system"
	fieldMessages    = "messages"
	fieldTools       = "tools"
	fieldToolChoice  = "tool_choice"
	fieldName        = "name"
	fieldDescription = "description"
	fieldInputSchema = "input_schema"
	fieldType        = "type"
	toolChoiceAuto   = "auto"

	fieldCacheControl = "cache_control"
	cacheEphemeral    = "ephemeral"

	fieldThinking     = "thinking"
	fieldOutputConfig = "output_config"
	fieldEffort       = "effort"
	fieldDisplay      = "display"
	fieldBudgetTokens = "budget_tokens"
	thinkingAdaptive  = "adaptive"
	thinkingEnabled   = "enabled"
	displaySummarized = "summarized"

	toolTypeFunction = "function"

	stopReasonMaxTokens = "max_tokens"
)

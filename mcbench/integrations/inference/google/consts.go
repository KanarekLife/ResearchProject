package google

// DefaultBaseURL is Google's own endpoint.
const DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Wire-format names of the generateContent API.
const (
	headerAPIKey = "x-goog-api-key"

	pathModels   = "/models/"
	methodSuffix = ":generateContent"

	roleUser  = "user"
	roleModel = "model"

	fieldContents          = "contents"
	fieldSystemInstruction = "systemInstruction"
	fieldGenerationConfig  = "generationConfig"
	fieldMaxOutputTokens   = "maxOutputTokens"
	fieldTemperature       = "temperature"
	fieldThinkingConfig    = "thinkingConfig"
	fieldThinkingLevel     = "thinkingLevel"
	fieldThinkingBudget    = "thinkingBudget"
	fieldTools             = "tools"
	fieldToolConfig        = "toolConfig"
	fieldFunctionDecls     = "functionDeclarations"
	fieldFunctionCalling   = "functionCallingConfig"
	fieldMode              = "mode"
	fieldParts             = "parts"
	fieldName              = "name"
	fieldDescription       = "description"
	fieldParameters        = "parameters"
	fieldType              = "type"
	fieldProperties        = "properties"
	fieldRequired          = "required"
	modeAuto               = "AUTO"
	typeObject             = "object"
	toolTypeFunction       = "function"

	finishReasonMaxTokens = "MAX_TOKENS"

	// idSep joins a tool call id: "name|signature". Gemini requires the
	// thought signature back on the call in later requests, and a function
	// response is matched by name.
	idSep = "|"
)

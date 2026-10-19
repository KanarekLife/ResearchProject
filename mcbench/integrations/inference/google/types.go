package google

import "encoding/json"

// Config is the endpoint and sampling settings, filled by the caller.
type Config struct {
	BaseURL     string // e.g. https://generativelanguage.googleapis.com/v1beta
	APIKey      string
	Model       string
	Temperature float64 // sent when >= 0
	MaxTokens   int
	Retries     int
	Headers     map[string]string // extra request headers; they override the defaults
	// ReasoningEffort is a thinking level ("low", "medium", "high"); ReasoningBudget
	// a thinking token cap, used only when no level is set. Empty or 0 leaves the
	// model's default.
	ReasoningEffort string
	ReasoningBudget int
}

type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text             string            `json:"text,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"`
	FunctionCall     *functionCall     `json:"functionCall,omitempty"`
	FunctionResponse *functionResponse `json:"functionResponse,omitempty"`
}

type functionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type functionResponse struct {
	Name     string          `json:"name"`
	Response json.RawMessage `json:"response"`
}

type response struct {
	Candidates []struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Usage struct {
		PromptTokens    int64 `json:"promptTokenCount"`
		CandidateTokens int64 `json:"candidatesTokenCount"`
		CachedTokens    int64 `json:"cachedContentTokenCount"`
		ThoughtsTokens  int64 `json:"thoughtsTokenCount"`
	} `json:"usageMetadata"`
}

package openairesponses

// Config is the endpoint and sampling settings, filled by the caller.
type Config struct {
	BaseURL     string // e.g. https://api.openai.com/v1
	APIKey      string
	Model       string
	Temperature float64 // sent when >= 0
	MaxTokens   int
	Retries     int
	// ReasoningEffort is a named thinking level (e.g. "high"); empty leaves the
	// model's default.
	ReasoningEffort string
}

// inputItem is one item of the input list: a message, a function call the
// model made, or the output of one. Function calls carry no item id, so the
// API does not ask for the reasoning item that went with them.
type inputItem struct {
	Type      string `json:"type,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
}

// response is the subset of a Responses reply we read.
type response struct {
	Status            string `json:"status"`
	IncompleteDetails struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
	Output []struct {
		Type      string `json:"type"`
		CallID    string `json:"call_id"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Content   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Summary []struct {
			Text string `json:"text"`
		} `json:"summary"`
	} `json:"output"`
	Usage struct {
		InputTokens        int64 `json:"input_tokens"`
		OutputTokens       int64 `json:"output_tokens"`
		InputTokensDetails struct {
			CachedTokens int64 `json:"cached_tokens"`
		} `json:"input_tokens_details"`
		OutputTokensDetails struct {
			ReasoningTokens int64 `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
	} `json:"usage"`
}

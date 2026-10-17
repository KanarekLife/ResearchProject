package anthropic

import "encoding/json"

// Config is the endpoint and sampling settings, filled by the caller.
type Config struct {
	BaseURL     string // e.g. https://api.anthropic.com/v1
	APIKey      string
	Model       string
	Temperature float64 // sent when >= 0
	MaxTokens   int
	Retries     int
}

type turn struct {
	Role    string  `json:"role"`
	Content []block `json:"content"`
}

type block struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

type response struct {
	Content    []block `json:"content"`
	StopReason string  `json:"stop_reason"`
	Usage      struct {
		InputTokens         int64 `json:"input_tokens"`
		OutputTokens        int64 `json:"output_tokens"`
		CacheReadTokens     int64 `json:"cache_read_input_tokens"`
		CacheCreationTokens int64 `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

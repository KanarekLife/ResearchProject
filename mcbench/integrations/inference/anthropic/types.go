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
	Headers     map[string]string // extra request headers; they override the defaults
	// ReasoningEffort is a named effort level (e.g. "high") sent with
	// adaptive thinking; empty leaves the model's default.
	ReasoningEffort string
	// ReasoningBudget is a thinking-token budget for models that take one
	// (at least 1024); used only when ReasoningEffort is empty.
	ReasoningBudget int
	// Stream asks for the reply as a stream, so that a gateway that cuts a
	// quiet request after a couple of minutes lets a slow reply through.
	Stream bool
}

type turn struct {
	Role    string `json:"role"`
	Content []any  `json:"content"` // block or json.RawMessage
}

type block struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"` // read only: thinking goes back raw
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

type response struct {
	Content    json.RawMessage `json:"content"`
	StopReason string          `json:"stop_reason"`
	Usage      usage           `json:"usage"`
}

type usage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheReadTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_input_tokens"`
}

// streamEvent is the part of a stream event we read.
type streamEvent struct {
	Type         string         `json:"type"`
	Index        int            `json:"index"`
	ContentBlock map[string]any `json:"content_block"`
	Message      struct {
		Usage usage `json:"usage"`
	} `json:"message"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		Thinking    string `json:"thinking"`
		Signature   string `json:"signature"`
		PartialJSON string `json:"partial_json"`
		StopReason  string `json:"stop_reason"`
	} `json:"delta"`
	Usage usage `json:"usage"`
}

// streamResponse is a response rebuilt from a stream, in the shape decodeReply
// reads.
type streamResponse struct {
	Content    json.RawMessage `json:"content"`
	StopReason string          `json:"stop_reason"`
	Usage      usage           `json:"usage"`
}

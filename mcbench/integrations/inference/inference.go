// Package inference defines the chat interface the model player depends on.
// It knows nothing about the game; implementations live in subpackages.
package inference

import "context"

// Client sends a conversation and returns the model's reply. tools are
// advertised for function calling when non-empty.
type Client interface {
	Complete(ctx context.Context, msgs []Message, tools []Tool) (Reply, error)
}

// Role is who sent a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is one chat message. Reasoning is the model's thinking, kept so the
// next request's prefix matches this one for prompt caching.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Reasoning  string     `json:"reasoning_content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall is one function call requested by the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Tool is a tool the model may call, as a JSON Schema.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Reply is a model response plus how it finished and what it cost.
type Reply struct {
	Message         Message
	Truncated       bool // the reply hit the token limit
	InputTokens     int64
	OutputTokens    int64
	CachedTokens    int64 // prefix-cache hits, when the provider reports them
	ReasoningTokens int64 // reasoning tokens, when the provider reports them
}

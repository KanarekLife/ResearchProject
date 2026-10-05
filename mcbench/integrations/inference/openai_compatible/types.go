package openaicompatible

import (
	"errors"
	"fmt"
	"time"

	"mcbench/integrations/inference"
)

// Config is the endpoint and sampling settings, filled by the caller.
type Config struct {
	BaseURL     string // e.g. http://localhost:1234/v1
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	Retries     int // extra attempts after a transient failure
}

// errTokenParam means the provider rejected max_tokens.
var errTokenParam = errors.New("provider wants max_completion_tokens")

// httpError is a non-200 response.
type httpError struct {
	status     int
	retryAfter time.Duration
	body       string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("chat API HTTP %d: %.300s", e.status, e.body)
}

// replyResponse is the subset of a chat-completions response we read.
type replyResponse struct {
	Choices []struct {
		Message      apiMessage `json:"message"`
		FinishReason string     `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens         int64 `json:"prompt_tokens"`
		CompletionTokens     int64 `json:"completion_tokens"`
		PromptCacheHitTokens int64 `json:"prompt_cache_hit_tokens"` // DeepSeek
		PromptTokensDetails  struct {
			CachedTokens int64 `json:"cached_tokens"`
		} `json:"prompt_tokens_details"` // OpenAI
		CompletionTokensDetails struct {
			ReasoningTokens int64 `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

// apiMessage decodes a provider message. Providers disagree on the field name
// for reasoning, so both are read.
type apiMessage struct {
	Role             string               `json:"role"`
	Content          string               `json:"content"`
	ReasoningContent string               `json:"reasoning_content"`
	Reasoning        string               `json:"reasoning"`
	ToolCalls        []inference.ToolCall `json:"tool_calls"`
}

func (m apiMessage) message() inference.Message {
	reasoning := m.ReasoningContent
	if reasoning == "" {
		reasoning = m.Reasoning
	}
	return inference.Message{
		Role:      inference.Role(m.Role),
		Content:   m.Content,
		Reasoning: reasoning,
		ToolCalls: m.ToolCalls,
	}
}

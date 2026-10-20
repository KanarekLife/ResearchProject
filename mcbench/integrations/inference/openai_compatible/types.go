package openaicompatible

import (
	"encoding/json"
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
	Temperature float64 // sent when >= 0; a negative value leaves the provider's default
	MaxTokens   int
	Retries     int               // extra attempts after a transient failure
	Headers     map[string]string // extra request headers; they override the defaults

	// ReasoningEffort is a named thinking level (e.g. "high") for hosted
	// models; empty leaves the provider default.
	ReasoningEffort string
	// ReasoningBudget caps thinking tokens on llama.cpp-style local servers;
	// 0 leaves the server default.
	ReasoningBudget int
	// Stream asks for the reply as a stream and assembles it. A gateway that
	// cuts a request after a couple of minutes lets a stream through, which
	// matters for slow reasoning on a long conversation.
	Stream bool
}

// errTokenParam means the provider rejected max_tokens.
var errTokenParam = errors.New("provider wants max_completion_tokens")

// errReasoningField means the provider rejected reasoning_content on messages.
var errReasoningField = errors.New("provider rejects reasoning_content on messages")

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
	Usage apiUsage `json:"usage"`
}

type apiUsage struct {
	PromptTokens         int64 `json:"prompt_tokens"`
	CompletionTokens     int64 `json:"completion_tokens"`
	PromptCacheHitTokens int64 `json:"prompt_cache_hit_tokens"` // DeepSeek
	PromptTokensDetails  struct {
		CachedTokens int64 `json:"cached_tokens"`
	} `json:"prompt_tokens_details"` // OpenAI
	CompletionTokensDetails struct {
		ReasoningTokens int64 `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

// apiMessage decodes a provider message. Providers disagree on the field name
// for reasoning, so both are read.
type apiMessage struct {
	Role             string               `json:"role"`
	Content          messageContent       `json:"content"`
	ReasoningContent string               `json:"reasoning_content"`
	Reasoning        string               `json:"reasoning"`
	ToolCalls        []inference.ToolCall `json:"tool_calls"`
}

func (m apiMessage) message() inference.Message {
	reasoning := m.ReasoningContent
	if reasoning == "" {
		reasoning = m.Reasoning
	}
	if reasoning == "" {
		reasoning = m.Content.thinking
	}
	return inference.Message{
		Role:      inference.Role(m.Role),
		Content:   m.Content.text,
		Reasoning: reasoning,
		ToolCalls: m.ToolCalls,
	}
}

// messageContent is a message's content: a string, or (Mistral) a list of
// "text" and "thinking" blocks. thinking holds the text of the thinking blocks.
type messageContent struct{ text, thinking string }

func (c *messageContent) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		c.text = s
		return nil
	}
	var blocks []struct {
		Type     string          `json:"type"`
		Text     string          `json:"text"`
		Thinking json.RawMessage `json:"thinking"`
	}
	if err := json.Unmarshal(b, &blocks); err != nil {
		return err
	}
	for _, blk := range blocks {
		switch blk.Type {
		case "text":
			c.text += blk.Text
		case "thinking":
			c.thinking += thinkingText(blk.Thinking)
		}
	}
	return nil
}

// thinkingText reads a thinking block's body: a string or a list of text parts.
func thinkingText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	json.Unmarshal(raw, &parts)
	var out string
	for _, p := range parts {
		out += p.Text
	}
	return out
}

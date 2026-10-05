package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mcbench/game"
)

// Chat is a client for an OpenAI-compatible chat completions API: local
// servers (LM Studio, Ollama, llama.cpp, vLLM) or hosted ones.
type Chat struct {
	BaseURL     string // e.g. http://localhost:1234/v1
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// Reasoning is the model's thinking (reasoning_content). It is kept in
	// the history so every request's prefix matches the previous one.
	Reasoning  string     `json:"reasoning_content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Reply struct {
	Message                   Message
	Truncated                 bool // the reply hit MaxTokens
	InputTokens, OutputTokens int64
}

var httpClient = &http.Client{Timeout: 10 * time.Minute}

// Complete sends the conversation and returns the model's reply. tools are
// offered for function calling when non-empty.
func (c Chat) Complete(ctx context.Context, msgs []Message, tools []game.Tool) (Reply, error) {
	body := map[string]any{"model": c.Model, "messages": msgs, "temperature": c.Temperature, "max_tokens": c.MaxTokens}
	if len(tools) > 0 {
		var ts []any
		for _, t := range tools {
			ts = append(ts, map[string]any{"type": "function", "function": t})
		}
		body["tools"], body["tool_choice"] = ts, "auto"
	}
	data, err := json.Marshal(body)
	if err != nil {
		return Reply{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return Reply{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Reply{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Reply{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Reply{}, fmt.Errorf("chat API HTTP %d: %.500s", resp.StatusCode, raw)
	}
	var r struct {
		Choices []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return Reply{}, fmt.Errorf("decoding chat response: %w", err)
	}
	if len(r.Choices) == 0 {
		return Reply{}, fmt.Errorf("chat response has no choices: %.500s", raw)
	}
	return Reply{Message: r.Choices[0].Message, Truncated: r.Choices[0].FinishReason == "length", InputTokens: r.Usage.PromptTokens, OutputTokens: r.Usage.CompletionTokens}, nil
}

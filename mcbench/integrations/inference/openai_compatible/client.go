// Package openaicompatible implements the inference client against an
// OpenAI-compatible chat-completions API.
package openaicompatible

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"mcbench/integrations/inference"
)

// requestTimeout bounds one HTTP request. Reasoning models can be slow.
const requestTimeout = 10 * time.Minute

// Client is an inference.Client backed by an OpenAI-compatible endpoint.
type Client struct{ cfg Config }

func New(cfg Config) *Client { return &Client{cfg: cfg} }

var httpClient = &http.Client{Timeout: requestTimeout}

// Complete sends the conversation and returns the model's reply, retrying
// transient failures.
func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	useMaxCompletion := false
	attempts := max(1, c.cfg.Retries+1)
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			c.backoff(ctx, attempt, last)
		}
		reply, err := c.request(ctx, msgs, tools, useMaxCompletion)
		if err == nil {
			return reply, nil
		}
		last = err
		slog.Warn("chat request failed", "attempt", attempt+1, "of", attempts, "error", err)
		if errors.Is(err, errTokenParam) {
			useMaxCompletion = true
			attempts++ // the parameter switch is not a retry
			continue
		}
		if !retryable(err) {
			return inference.Reply{}, err
		}
	}
	return inference.Reply{}, fmt.Errorf("chat request failed after %d attempts: %w", attempts, last)
}

func (c *Client) request(ctx context.Context, msgs []inference.Message, tools []inference.Tool, useMaxCompletion bool) (inference.Reply, error) {
	data, err := json.Marshal(c.body(msgs, tools, useMaxCompletion))
	if err != nil {
		return inference.Reply{}, err
	}
	resp, err := c.post(ctx, data)
	if err != nil {
		return inference.Reply{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return inference.Reply{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return inference.Reply{}, statusError(resp, raw, useMaxCompletion)
	}
	return decodeReply(raw)
}

func (c *Client) body(msgs []inference.Message, tools []inference.Tool, useMaxCompletion bool) map[string]any {
	body := map[string]any{fieldModel: c.cfg.Model, fieldMessages: msgs}
	tokenField := fieldMaxTokens
	if useMaxCompletion {
		tokenField = fieldMaxCompletionTokens
	}
	body[tokenField] = c.cfg.MaxTokens
	if c.cfg.Temperature > 0 {
		body[fieldTemperature] = c.cfg.Temperature
	}
	if len(tools) > 0 {
		fns := make([]any, len(tools))
		for i, t := range tools {
			fns[i] = map[string]any{fieldType: fieldFunction, fieldFunction: t}
		}
		body[fieldTools], body[fieldToolChoice] = fns, toolChoiceAuto
	}
	return body
}

func (c *Client) post(ctx context.Context, data []byte) (*http.Response, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathChatCompletions
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, mimeJSON)
	if c.cfg.APIKey != "" {
		req.Header.Set(headerAuthorization, bearerPrefix+c.cfg.APIKey)
	}
	return httpClient.Do(req)
}

func decodeReply(raw []byte) (inference.Reply, error) {
	var r replyResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return inference.Reply{}, fmt.Errorf("decoding chat response: %w", err)
	}
	if len(r.Choices) == 0 {
		return inference.Reply{}, fmt.Errorf("chat response has no choices: %.300s", raw)
	}
	return inference.Reply{
		Message:         r.Choices[0].Message.message(),
		Truncated:       r.Choices[0].FinishReason == finishReasonLength,
		InputTokens:     r.Usage.PromptTokens,
		OutputTokens:    r.Usage.CompletionTokens,
		CachedTokens:    max(r.Usage.PromptTokensDetails.CachedTokens, r.Usage.PromptCacheHitTokens),
		ReasoningTokens: r.Usage.CompletionTokensDetails.ReasoningTokens,
	}, nil
}

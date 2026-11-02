// Package openaicompatible implements the inference client against an
// OpenAI-compatible chat-completions API.
package openaicompatible

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"sync/atomic"

	"mcbench/integrations/inference"
)

// Client is an inference.Client backed by an OpenAI-compatible endpoint.
type Client struct {
	cfg Config
	// dropReasoning is set once the provider rejects reasoning_content on
	// messages; later requests then leave the model's reasoning out.
	dropReasoning atomic.Bool
}

func New(cfg Config) *Client { return &Client{cfg: cfg} }

// Complete sends the conversation and returns the model's reply, retrying
// transient failures and waiting out rate limits. A provider that rejects
// max_tokens or reasoning_content is asked again at once without it; that
// switch is not a retry.
func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	useMaxCompletion := false
	return inference.Retry(ctx, c.cfg.Retries, c.cfg.RateLimitWait, func() (inference.Reply, error) {
		for {
			reply, err := c.request(ctx, msgs, tools, useMaxCompletion)
			switch {
			case errors.Is(err, errTokenParam) && !useMaxCompletion:
				useMaxCompletion = true
			case errors.Is(err, errReasoningField) && !c.dropReasoning.Load():
				c.dropReasoning.Store(true)
			default:
				return reply, err
			}
			slog.Warn("chat request rejected a parameter, asking again", "error", err)
		}
	})
}

func (c *Client) request(ctx context.Context, msgs []inference.Message, tools []inference.Tool, useMaxCompletion bool) (inference.Reply, error) {
	if c.dropReasoning.Load() {
		msgs = withoutReasoning(msgs)
	}
	data, err := json.Marshal(c.body(msgs, tools, useMaxCompletion))
	if err != nil {
		return inference.Reply{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathChatCompletions
	headers := map[string]string{}
	if c.cfg.APIKey != "" {
		headers[headerAuthorization] = bearerPrefix + c.cfg.APIKey
	}
	maps.Copy(headers, inference.ExpandHeaders(ctx, c.cfg.Headers))
	var reply inference.Reply
	if c.cfg.Stream {
		var s stream
		if err = inference.PostSSE(ctx, url, headers, data, s.add); err == nil {
			reply, err = s.reply()
		}
	} else {
		var raw []byte
		if raw, err = inference.PostJSON(ctx, url, headers, data); err == nil {
			reply, err = decodeReply(raw)
		}
	}
	return reply, rejected(err, msgs, useMaxCompletion)
}

func (c *Client) body(msgs []inference.Message, tools []inference.Tool, useMaxCompletion bool) map[string]any {
	body := map[string]any{fieldModel: c.cfg.Model, fieldMessages: msgs}
	tokenField := fieldMaxTokens
	if useMaxCompletion {
		tokenField = fieldMaxCompletionTokens
	}
	body[tokenField] = c.cfg.MaxTokens
	if c.cfg.Temperature >= 0 {
		body[fieldTemperature] = c.cfg.Temperature
	}
	if c.cfg.ReasoningEffort != "" {
		body[fieldReasoningEffort] = c.cfg.ReasoningEffort
	}
	if c.cfg.ReasoningBudget > 0 {
		// llama.cpp reads reasoning_budget_tokens; older builds only the
		// thinking_budget_tokens alias.
		body[fieldReasoningBudget] = c.cfg.ReasoningBudget
		body[fieldThinkingBudget] = c.cfg.ReasoningBudget
	}
	if c.cfg.Stream {
		body[fieldStream] = true
		body[fieldStreamOptions] = map[string]any{fieldIncludeUsage: true}
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

func hasReasoning(msgs []inference.Message) bool {
	for _, m := range msgs {
		if m.Reasoning != "" {
			return true
		}
	}
	return false
}

// withoutReasoning copies msgs without the model's reasoning.
func withoutReasoning(msgs []inference.Message) []inference.Message {
	out := make([]inference.Message, len(msgs))
	for i, m := range msgs {
		m.Reasoning = ""
		out[i] = m
	}
	return out
}

// Package openairesponses implements the inference client against the OpenAI
// Responses API, for models that are not served over chat completions. The
// conversation is sent whole on every request (stateless).
package openairesponses

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"mcbench/integrations/inference"
)

// Client is an inference.Client backed by the Responses API.
type Client struct{ cfg Config }

func New(cfg Config) *Client { return &Client{cfg: cfg} }

func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	body, err := json.Marshal(c.body(msgs, tools))
	if err != nil {
		return inference.Reply{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathResponses
	headers := map[string]string{}
	if c.cfg.APIKey != "" {
		headers[headerAuthorization] = bearerPrefix + c.cfg.APIKey
	}
	maps.Copy(headers, inference.ExpandHeaders(ctx, c.cfg.Headers))
	return inference.Retry(ctx, c.cfg.Retries, c.cfg.RateLimitWait, func() (inference.Reply, error) {
		if c.cfg.Stream {
			return c.stream(ctx, url, headers, body)
		}
		raw, err := inference.PostJSON(ctx, url, headers, body)
		if err != nil {
			return inference.Reply{}, err
		}
		return decodeReply(raw)
	})
}

// stream reads a streamed reply. Its last event carries the whole response, so
// the deltas before it are not needed.
func (c *Client) stream(ctx context.Context, url string, headers map[string]string, body []byte) (inference.Reply, error) {
	var final json.RawMessage
	err := inference.PostSSE(ctx, url, headers, body, func(_ string, data []byte) error {
		var ev streamEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil // [DONE] and the like
		}
		switch ev.Type {
		case eventCompleted, eventIncomplete:
			final = ev.Response
		case eventFailed, eventError:
			return fmt.Errorf("responses stream failed: %.300s", data)
		}
		return nil
	})
	if err != nil {
		return inference.Reply{}, err
	}
	if final == nil {
		return inference.Reply{}, fmt.Errorf("responses stream ended before the response completed")
	}
	return decodeReply(final)
}

func (c *Client) body(msgs []inference.Message, tools []inference.Tool) map[string]any {
	// Nothing is stored server side, so the reasoning comes back encrypted
	// and is returned with the reply it belongs to on every later request.
	body := map[string]any{
		fieldModel: c.cfg.Model, fieldMaxOutputTokens: c.cfg.MaxTokens, fieldStore: false,
		fieldInclude: []string{includeEncryptedReasoning},
	}
	if c.cfg.Stream {
		body[fieldStream] = true
	}
	if c.cfg.Temperature >= 0 {
		body[fieldTemperature] = c.cfg.Temperature
	}
	if c.cfg.ReasoningEffort != "" {
		body[fieldReasoning] = map[string]any{fieldEffort: c.cfg.ReasoningEffort}
	}
	var items []any // inputItem, or an output item sent back verbatim
	for _, m := range msgs {
		switch m.Role {
		case inference.RoleSystem:
			body[fieldInstructions] = m.Content
		case inference.RoleUser:
			items = append(items, inputItem{Role: roleUser, Content: m.Content})
		case inference.RoleAssistant:
			if len(m.Native) > 0 {
				var out []json.RawMessage
				json.Unmarshal(m.Native, &out) // set by decodeReply, so always valid
				for _, it := range out {
					items = append(items, it)
				}
				continue
			}
			if m.Content != "" {
				items = append(items, inputItem{Role: roleAssistant, Content: m.Content})
			}
			for _, tc := range m.ToolCalls {
				args := tc.Function.Arguments
				if strings.TrimSpace(args) == "" {
					args = "{}"
				}
				items = append(items, inputItem{Type: itemFunctionCall, CallID: tc.ID, Name: tc.Function.Name, Arguments: args})
			}
		case inference.RoleTool:
			items = append(items, inputItem{Type: itemFunctionCallOutput, CallID: m.ToolCallID, Output: m.Content})
		}
	}
	body[fieldInput] = items
	if len(tools) > 0 {
		defs := make([]any, len(tools))
		for i, t := range tools {
			defs[i] = map[string]any{fieldType: toolTypeFunction, fieldName: t.Name, fieldDescription: t.Description, fieldParameters: t.Parameters}
		}
		body[fieldTools], body[fieldToolChoice] = defs, toolChoiceAuto
	}
	return body
}

func decodeReply(raw []byte) (inference.Reply, error) {
	var r response
	if err := json.Unmarshal(raw, &r); err != nil {
		return inference.Reply{}, fmt.Errorf("decoding responses reply: %w", err)
	}
	var output []outputItem
	if len(r.Output) > 0 {
		if err := json.Unmarshal(r.Output, &output); err != nil {
			return inference.Reply{}, fmt.Errorf("decoding responses output: %w", err)
		}
	}
	msg := inference.Message{Role: inference.RoleAssistant}
	if len(output) > 0 {
		// The output goes back verbatim: its reasoning items carry the
		// encrypted reasoning, and its function calls the ids that tie them
		// to it.
		msg.Native = r.Output
	}
	for _, it := range output {
		switch it.Type {
		case itemMessage:
			for _, c := range it.Content {
				if c.Type == contentOutputText {
					msg.Content += c.Text
				}
			}
		case itemFunctionCall:
			var tc inference.ToolCall
			tc.ID, tc.Type = it.CallID, toolTypeFunction
			tc.Function.Name, tc.Function.Arguments = it.Name, it.Arguments
			msg.ToolCalls = append(msg.ToolCalls, tc)
		case itemReasoning:
			for _, s := range it.Summary {
				msg.Reasoning += s.Text
			}
		}
	}
	return inference.Reply{
		Message:         msg,
		Truncated:       r.Status == statusIncomplete && r.IncompleteDetails.Reason == reasonMaxOutputTokens,
		InputTokens:     r.Usage.InputTokens,
		OutputTokens:    r.Usage.OutputTokens,
		CachedTokens:    r.Usage.InputTokensDetails.CachedTokens,
		ReasoningTokens: r.Usage.OutputTokensDetails.ReasoningTokens,
	}, nil
}

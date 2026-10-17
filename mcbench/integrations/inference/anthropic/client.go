// Package anthropic implements the inference client against the Anthropic
// Messages API (api.anthropic.com, or any server that proxies it). It sends
// no thinking or cache settings.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mcbench/integrations/inference"
)

// Client is an inference.Client backed by the Messages API.
type Client struct{ cfg Config }

func New(cfg Config) *Client { return &Client{cfg: cfg} }

func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	body, err := json.Marshal(c.body(msgs, tools))
	if err != nil {
		return inference.Reply{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathMessages
	headers := map[string]string{headerAPIKey: c.cfg.APIKey, headerVersion: apiVersion}
	return inference.Retry(ctx, c.cfg.Retries, func() (inference.Reply, error) {
		raw, err := inference.PostJSON(ctx, url, headers, body)
		if err != nil {
			return inference.Reply{}, err
		}
		return decodeReply(raw)
	})
}

func (c *Client) body(msgs []inference.Message, tools []inference.Tool) map[string]any {
	body := map[string]any{fieldModel: c.cfg.Model, fieldMaxTokens: c.cfg.MaxTokens}
	if c.cfg.Temperature >= 0 {
		body[fieldTemperature] = c.cfg.Temperature
	}
	var turns []turn
	for _, m := range msgs {
		switch m.Role {
		case inference.RoleSystem:
			body[fieldSystem] = m.Content
		case inference.RoleUser:
			turns = appendTurn(turns, roleUser, block{Type: blockText, Text: m.Content})
		case inference.RoleAssistant:
			if len(m.Native) > 0 {
				turns = appendTurn(turns, roleAssistant, nativeBlocks(m.Native)...)
				continue
			}
			var blocks []any
			if m.Content != "" {
				blocks = append(blocks, block{Type: blockText, Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				args := tc.Function.Arguments
				if strings.TrimSpace(args) == "" {
					args = "{}"
				}
				blocks = append(blocks, block{Type: blockToolUse, ID: tc.ID, Name: tc.Function.Name, Input: json.RawMessage(args)})
			}
			if len(blocks) == 0 {
				blocks = append(blocks, block{Type: blockText, Text: " "}) // a turn needs at least one block
			}
			turns = appendTurn(turns, roleAssistant, blocks...)
		case inference.RoleTool:
			turns = appendTurn(turns, roleUser, block{Type: blockToolResult, ToolUseID: m.ToolCallID, Content: m.Content})
		}
	}
	body[fieldMessages] = turns
	if len(tools) > 0 {
		defs := make([]any, len(tools))
		for i, t := range tools {
			defs[i] = map[string]any{fieldName: t.Name, fieldDescription: t.Description, fieldInputSchema: t.Parameters}
		}
		body[fieldTools] = defs
		body[fieldToolChoice] = map[string]any{fieldType: toolChoiceAuto}
	}
	return body
}

// appendTurn adds blocks to the last turn when it has the same role: the
// tool results of one assistant turn, and any user text after them, share one
// user turn.
func appendTurn(ts []turn, role string, blocks ...any) []turn {
	if n := len(ts); n > 0 && ts[n-1].Role == role {
		ts[n-1].Content = append(ts[n-1].Content, blocks...)
		return ts
	}
	return append(ts, turn{Role: role, Content: blocks})
}

// nativeBlocks splits a reply's raw content array (set by decodeReply, so
// always valid) into its blocks, unchanged.
func nativeBlocks(native json.RawMessage) []any {
	var raw []json.RawMessage
	json.Unmarshal(native, &raw)
	out := make([]any, len(raw))
	for i, b := range raw {
		out[i] = b
	}
	return out
}

func decodeReply(raw []byte) (inference.Reply, error) {
	var r response
	if err := json.Unmarshal(raw, &r); err != nil {
		return inference.Reply{}, fmt.Errorf("decoding messages response: %w", err)
	}
	var blocks []block
	if len(r.Content) > 0 {
		if err := json.Unmarshal(r.Content, &blocks); err != nil {
			return inference.Reply{}, fmt.Errorf("decoding messages content: %w", err)
		}
	}
	msg := inference.Message{Role: inference.RoleAssistant}
	if len(blocks) > 0 {
		// The content goes back verbatim: thinking blocks must keep their
		// signatures, and the text and tool calls their order around them.
		msg.Native = r.Content
	}
	for _, b := range blocks {
		switch b.Type {
		case blockThinking:
			msg.Reasoning += b.Thinking
		case "text":
			msg.Content += b.Text
		case "tool_use":
			var tc inference.ToolCall
			tc.ID, tc.Type = b.ID, "function"
			tc.Function.Name, tc.Function.Arguments = b.Name, string(b.Input)
			msg.ToolCalls = append(msg.ToolCalls, tc)
		}
	}
	u := r.Usage
	return inference.Reply{
		Message:      msg,
		Truncated:    r.StopReason == "max_tokens",
		InputTokens:  u.InputTokens + u.CacheReadTokens + u.CacheCreationTokens, // input_tokens excludes cached ones
		OutputTokens: u.OutputTokens,
		CachedTokens: u.CacheReadTokens,
	}, nil
}

// Package google implements the inference client against the Gemini
// generateContent API (Google AI Studio, or any server that proxies it).
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"mcbench/integrations/inference"
)

// Client is an inference.Client backed by the generateContent API.
type Client struct{ cfg Config }

func New(cfg Config) *Client { return &Client{cfg: cfg} }

func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	body, err := json.Marshal(c.body(msgs, tools))
	if err != nil {
		return inference.Reply{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathModels + c.cfg.Model + methodSuffix
	headers := map[string]string{headerAPIKey: c.cfg.APIKey}
	maps.Copy(headers, c.cfg.Headers)
	return inference.Retry(ctx, c.cfg.Retries, func() (inference.Reply, error) {
		raw, err := inference.PostJSON(ctx, url, headers, body)
		if err != nil {
			return inference.Reply{}, err
		}
		return decodeReply(raw)
	})
}

func (c *Client) body(msgs []inference.Message, tools []inference.Tool) map[string]any {
	gen := map[string]any{fieldMaxOutputTokens: c.cfg.MaxTokens}
	if c.cfg.Temperature >= 0 {
		gen[fieldTemperature] = c.cfg.Temperature
	}
	// The API rejects a thinking level together with a thinking budget, so the
	// level wins when both are set.
	think := map[string]any{}
	if c.cfg.ReasoningEffort != "" {
		think[fieldThinkingLevel] = c.cfg.ReasoningEffort
	} else if c.cfg.ReasoningBudget > 0 {
		think[fieldThinkingBudget] = c.cfg.ReasoningBudget
	}
	if len(think) > 0 {
		gen[fieldThinkingConfig] = think
	}
	body := map[string]any{fieldGenerationConfig: gen}
	var contents []content
	for _, m := range msgs {
		switch m.Role {
		case inference.RoleSystem:
			body[fieldSystemInstruction] = map[string]any{fieldParts: []part{{Text: m.Content}}}
		case inference.RoleUser:
			contents = appendContent(contents, roleUser, part{Text: m.Content})
		case inference.RoleAssistant:
			var parts []part
			if m.Content != "" {
				parts = append(parts, part{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				name, sig, _ := strings.Cut(tc.ID, idSep)
				parts = append(parts, part{
					FunctionCall: &functionCall{Name: name, Args: json.RawMessage(orEmptyObject(tc.Function.Arguments))}, ThoughtSignature: sig,
				})
			}
			if len(parts) == 0 {
				parts = append(parts, part{Text: " "}) // a content needs at least one part
			}
			contents = appendContent(contents, roleModel, parts...)
		case inference.RoleTool:
			name, _, _ := strings.Cut(m.ToolCallID, idSep)
			contents = appendContent(contents, roleUser, part{FunctionResponse: &functionResponse{Name: name, Response: asObject(m.Content)}})
		}
	}
	body[fieldContents] = contents
	if len(tools) > 0 {
		decls := make([]any, len(tools))
		for i, t := range tools {
			decls[i] = declaration(t)
		}
		body[fieldTools] = []any{map[string]any{fieldFunctionDecls: decls}}
		body[fieldToolConfig] = map[string]any{fieldFunctionCalling: map[string]any{fieldMode: modeAuto}}
	}
	return body
}

// appendContent adds parts to the last content when it has the same role: the
// API wants all responses to one turn's calls in a single content.
func appendContent(cs []content, role string, parts ...part) []content {
	if n := len(cs); n > 0 && cs[n-1].Role == role {
		cs[n-1].Parts = append(cs[n-1].Parts, parts...)
		return cs
	}
	return append(cs, content{Role: role, Parts: parts})
}

// declaration is a tool as a function declaration. The API accepts a subset of
// JSON Schema: no additionalProperties and no empty required list, and a tool
// without parameters has no schema at all.
func declaration(t inference.Tool) map[string]any {
	d := map[string]any{fieldName: t.Name, fieldDescription: t.Description}
	if props, _ := t.Parameters[fieldProperties].(map[string]any); len(props) > 0 {
		schema := map[string]any{fieldType: typeObject, fieldProperties: props}
		if req, _ := t.Parameters[fieldRequired].([]string); len(req) > 0 {
			schema[fieldRequired] = req
		}
		d[fieldParameters] = schema
	}
	return d
}

func orEmptyObject(s string) string {
	if strings.TrimSpace(s) == "" {
		return "{}"
	}
	return s
}

// asObject wraps a tool result for functionResponse.response, which must be an
// object. Our results already are JSON objects.
func asObject(s string) json.RawMessage {
	var obj map[string]json.RawMessage
	if json.Unmarshal([]byte(s), &obj) == nil && obj != nil {
		return json.RawMessage(s)
	}
	b, _ := json.Marshal(map[string]string{"result": s})
	return b
}

func decodeReply(raw []byte) (inference.Reply, error) {
	var r response
	if err := json.Unmarshal(raw, &r); err != nil {
		return inference.Reply{}, fmt.Errorf("decoding generateContent response: %w", err)
	}
	if len(r.Candidates) == 0 {
		return inference.Reply{}, fmt.Errorf("response has no candidates (block reason %q): %.300s", r.PromptFeedback.BlockReason, raw)
	}
	cand := r.Candidates[0]
	msg := inference.Message{Role: inference.RoleAssistant}
	for _, p := range cand.Content.Parts {
		switch {
		case p.FunctionCall != nil:
			var tc inference.ToolCall
			tc.ID, tc.Type = p.FunctionCall.Name+idSep+p.ThoughtSignature, "function"
			tc.Function.Name, tc.Function.Arguments = p.FunctionCall.Name, string(p.FunctionCall.Args)
			msg.ToolCalls = append(msg.ToolCalls, tc)
		case p.Thought:
			msg.Reasoning += p.Text
		default:
			msg.Content += p.Text
		}
	}
	return inference.Reply{
		Message:         msg,
		Truncated:       cand.FinishReason == "MAX_TOKENS",
		InputTokens:     r.Usage.PromptTokens,
		OutputTokens:    r.Usage.CandidateTokens + r.Usage.ThoughtsTokens, // thoughts are billed as output
		CachedTokens:    r.Usage.CachedTokens,
		ReasoningTokens: r.Usage.ThoughtsTokens,
	}, nil
}

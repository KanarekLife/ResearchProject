package openaicompatible

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"mcbench/integrations/inference"
)

const (
	sseDataPrefix = "data:"
	sseDone       = "[DONE]"
	// maxStreamLine bounds one event; a chunk is small, but tool arguments can be long.
	maxStreamLine = 1 << 24
)

// streamChunk is one event of a streamed chat completion.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Reasoning        string `json:"reasoning"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *apiUsage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// decodeStream assembles a streamed reply into the reply a plain request returns.
func decodeStream(r io.Reader) (inference.Reply, error) {
	var (
		content, reasoning strings.Builder
		finish             string
		usage              apiUsage
		calls              []*inference.ToolCall
	)
	sc := bufio.NewScanner(r)
	sc.Buffer(nil, maxStreamLine)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), sseDataPrefix)
		if data = strings.TrimSpace(data); !ok || data == "" {
			continue // a comment or keep-alive line
		}
		if data == sseDone {
			break
		}
		var ch streamChunk
		if err := json.Unmarshal([]byte(data), &ch); err != nil {
			return inference.Reply{}, fmt.Errorf("decoding stream event: %w", err)
		}
		if ch.Error != nil {
			return inference.Reply{}, fmt.Errorf("stream failed: %s", ch.Error.Message)
		}
		if ch.Usage != nil {
			usage = *ch.Usage
		}
		for _, c := range ch.Choices {
			content.WriteString(c.Delta.Content)
			reasoning.WriteString(c.Delta.ReasoningContent + c.Delta.Reasoning)
			for _, tc := range c.Delta.ToolCalls {
				for len(calls) <= tc.Index {
					calls = append(calls, &inference.ToolCall{})
				}
				call := calls[tc.Index]
				if tc.ID != "" {
					call.ID = tc.ID
				}
				if tc.Type != "" {
					call.Type = tc.Type
				}
				call.Function.Name += tc.Function.Name
				call.Function.Arguments += tc.Function.Arguments
			}
			if c.FinishReason != "" {
				finish = c.FinishReason
			}
		}
	}
	if err := sc.Err(); err != nil {
		return inference.Reply{}, fmt.Errorf("reading stream: %w", err)
	}
	if finish == "" {
		return inference.Reply{}, fmt.Errorf("stream ended before the reply finished")
	}
	msg := inference.Message{Role: inference.RoleAssistant, Content: content.String(), Reasoning: reasoning.String()}
	for _, c := range calls {
		msg.ToolCalls = append(msg.ToolCalls, *c)
	}
	return inference.Reply{
		Message:         msg,
		Truncated:       finish == finishReasonLength,
		InputTokens:     usage.PromptTokens,
		OutputTokens:    usage.CompletionTokens,
		CachedTokens:    max(usage.PromptTokensDetails.CachedTokens, usage.PromptCacheHitTokens),
		ReasoningTokens: usage.CompletionTokensDetails.ReasoningTokens,
	}, nil
}

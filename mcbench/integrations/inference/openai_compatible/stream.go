package openaicompatible

import (
	"encoding/json"
	"fmt"
	"strings"

	"mcbench/integrations/inference"
)

// sseDone is the event that ends a chat-completions stream.
const sseDone = "[DONE]"

// stream assembles a streamed reply, one event at a time, into the reply a
// plain request returns.
type stream struct {
	content, reasoning strings.Builder
	finish             string
	usage              apiUsage
	calls              []*inference.ToolCall
}

func (s *stream) add(_ string, raw []byte) error {
	data := strings.TrimSpace(string(raw))
	if data == "" || data == sseDone {
		return nil
	}
	var ch streamChunk
	if err := json.Unmarshal([]byte(data), &ch); err != nil {
		return fmt.Errorf("decoding stream event: %w", err)
	}
	if ch.Error != nil {
		return fmt.Errorf("stream failed: %s", ch.Error.Message)
	}
	if ch.Usage != nil {
		s.usage = *ch.Usage
	}
	for _, c := range ch.Choices {
		s.content.WriteString(c.Delta.Content)
		s.reasoning.WriteString(c.Delta.ReasoningContent + c.Delta.Reasoning)
		for _, tc := range c.Delta.ToolCalls {
			for len(s.calls) <= tc.Index {
				s.calls = append(s.calls, &inference.ToolCall{})
			}
			call := s.calls[tc.Index]
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
			s.finish = c.FinishReason
		}
	}
	return nil
}

func (s *stream) reply() (inference.Reply, error) {
	if s.finish == "" {
		return inference.Reply{}, fmt.Errorf("stream ended before the reply finished")
	}
	msg := inference.Message{Role: inference.RoleAssistant, Content: s.content.String(), Reasoning: s.reasoning.String()}
	for _, c := range s.calls {
		msg.ToolCalls = append(msg.ToolCalls, *c)
	}
	return inference.Reply{
		Message:         msg,
		Truncated:       s.finish == finishReasonLength,
		InputTokens:     s.usage.PromptTokens,
		OutputTokens:    s.usage.CompletionTokens,
		CachedTokens:    max(s.usage.PromptTokensDetails.CachedTokens, s.usage.PromptCacheHitTokens),
		ReasoningTokens: s.usage.CompletionTokensDetails.ReasoningTokens,
	}, nil
}

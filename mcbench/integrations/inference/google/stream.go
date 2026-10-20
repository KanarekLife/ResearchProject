package google

import (
	"context"
	"encoding/json"
	"fmt"

	"mcbench/integrations/inference"
)

// stream reads a streamed reply: every event is a response chunk holding the
// next parts of the one candidate. The parts are joined into the response a
// plain request returns, and the finish reason and usage are the last ones the
// stream reports.
func stream(ctx context.Context, url string, headers map[string]string, body []byte) (inference.Reply, error) {
	var (
		merged response
		cand   struct {
			parts  []part
			finish string
			any    bool
		}
	)
	err := inference.PostSSE(ctx, url, headers, body, func(_ string, data []byte) error {
		var chunk response
		if err := json.Unmarshal(data, &chunk); err != nil {
			return fmt.Errorf("decoding generateContent stream event: %w", err)
		}
		if chunk.PromptFeedback.BlockReason != "" {
			merged.PromptFeedback = chunk.PromptFeedback
		}
		if chunk.Usage != (response{}).Usage {
			merged.Usage = chunk.Usage
		}
		if len(chunk.Candidates) > 0 {
			cand.any = true
			c := chunk.Candidates[0]
			for _, p := range c.Content.Parts {
				cand.parts = appendPart(cand.parts, p)
			}
			if c.FinishReason != "" {
				cand.finish = c.FinishReason
			}
		}
		return nil
	})
	if err != nil {
		return inference.Reply{}, err
	}
	if cand.any && cand.finish == "" {
		return inference.Reply{}, fmt.Errorf("generateContent stream ended before the reply finished")
	}
	if cand.any {
		merged.Candidates = append(merged.Candidates, struct {
			Content      content `json:"content"`
			FinishReason string  `json:"finishReason"`
		}{Content: content{Role: roleModel, Parts: cand.parts}, FinishReason: cand.finish})
	}
	return replyFrom(merged, nil)
}

// appendPart adds p after parts, joining it to the last part when both are
// plain text of the same kind (thought or answer), since a stream splits text
// at arbitrary points.
func appendPart(parts []part, p part) []part {
	if n := len(parts); n > 0 && p.FunctionCall == nil && parts[n-1].FunctionCall == nil && parts[n-1].Thought == p.Thought {
		parts[n-1].Text += p.Text
		if p.ThoughtSignature != "" {
			parts[n-1].ThoughtSignature = p.ThoughtSignature
		}
		return parts
	}
	return append(parts, p)
}

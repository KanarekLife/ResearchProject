package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mcbench/integrations/inference"
)

// Stream event types.
const (
	eventMessageStart = "message_start"
	eventBlockStart   = "content_block_start"
	eventBlockDelta   = "content_block_delta"
	eventMessageDelta = "message_delta"
	eventError        = "error"

	deltaText      = "text_delta"
	deltaThinking  = "thinking_delta"
	deltaSignature = "signature_delta"
	deltaInputJSON = "input_json_delta"
)

// stream reads a streamed reply and rebuilds the response a plain request
// returns: the content blocks complete (thinking with its signature, tool input
// as one object), the stop reason and the usage. The result then goes through
// decodeReply, so the reply is the same either way.
func stream(ctx context.Context, url string, headers map[string]string, body []byte) (inference.Reply, error) {
	var (
		blocks   []map[string]any
		toolJSON = map[int]*strings.Builder{} // partial tool input by block index
		resp     = streamResponse{}
		finished bool
	)
	err := inference.PostSSE(ctx, url, headers, body, func(_ string, data []byte) error {
		var ev streamEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil
		}
		switch ev.Type {
		case eventMessageStart:
			resp.Usage = ev.Message.Usage
		case eventBlockStart:
			for len(blocks) <= ev.Index {
				blocks = append(blocks, nil)
			}
			blocks[ev.Index] = ev.ContentBlock
			if ev.ContentBlock[fieldType] == blockToolUse {
				toolJSON[ev.Index] = &strings.Builder{}
			}
		case eventBlockDelta:
			if ev.Index >= len(blocks) || blocks[ev.Index] == nil {
				return fmt.Errorf("stream delta for unknown block %d", ev.Index)
			}
			b := blocks[ev.Index]
			switch ev.Delta.Type {
			case deltaText:
				appendString(b, blockText, ev.Delta.Text)
			case deltaThinking:
				appendString(b, blockThinking, ev.Delta.Thinking)
			case deltaSignature:
				b["signature"] = ev.Delta.Signature
			case deltaInputJSON:
				toolJSON[ev.Index].WriteString(ev.Delta.PartialJSON)
			}
		case eventMessageDelta:
			resp.StopReason = ev.Delta.StopReason
			resp.Usage.OutputTokens = ev.Usage.OutputTokens
		case eventError:
			return fmt.Errorf("messages stream failed: %.300s", data)
		case "message_stop":
			finished = true
		}
		return nil
	})
	if err != nil {
		return inference.Reply{}, err
	}
	if !finished {
		return inference.Reply{}, fmt.Errorf("messages stream ended before the message finished")
	}
	for i, b := range blocks {
		if sb, ok := toolJSON[i]; ok {
			input := json.RawMessage(sb.String())
			if strings.TrimSpace(sb.String()) == "" {
				input = json.RawMessage("{}")
			}
			b["input"] = input
		}
	}
	content, err := json.Marshal(blocks)
	if err != nil {
		return inference.Reply{}, err
	}
	resp.Content = content
	raw, err := json.Marshal(resp)
	if err != nil {
		return inference.Reply{}, err
	}
	return decodeReply(raw)
}

// appendString adds s to the string field key of a block.
func appendString(b map[string]any, key, s string) {
	old, _ := b[key].(string)
	b[key] = old + s
}

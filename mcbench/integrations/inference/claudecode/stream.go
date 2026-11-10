package claudecode

import (
	"encoding/json"
	"strings"

	"mcbench/integrations/inference"
)

// Claude Code's stream-json output: one event per line.
const (
	eventStream       = "stream_event" // a raw Messages API stream event
	eventAssistant    = "assistant"    // a finished content block of a reply
	eventResult       = "result"       // the end of a turn
	eventRateLimit    = "rate_limit_event"
	eventSystem       = "system"
	subtypeCompact    = "compact_boundary"
	rateLimitRejected = "rejected"

	streamMessageStart = "message_start"
	streamBlockStart   = "content_block_start"
	streamMessageDelta = "message_delta"
	streamMessageStop  = "message_stop"

	blockText     = "text"
	blockThinking = "thinking"
	blockToolUse  = "tool_use"
	stopMaxTokens = "max_tokens"
)

type event struct {
	Type      string        `json:"type"`
	Subtype   string        `json:"subtype"`
	Event     streamEvent   `json:"event"`
	Message   apiMessage    `json:"message"`
	IsError   bool          `json:"is_error"`
	Result    string        `json:"result"`
	RateLimit rateLimitInfo `json:"rate_limit_info"`
}

type rateLimitInfo struct {
	Status   string `json:"status"`
	Type     string `json:"rateLimitType"`
	ResetsAt int64  `json:"resetsAt"`
}

type streamEvent struct {
	Type    string `json:"type"`
	Message struct {
		Usage usage `json:"usage"`
	} `json:"message"`
	Delta struct {
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
	Usage usage `json:"usage"`
}

type apiMessage struct {
	Content []block `json:"content"`
}

type block struct {
	Type     string          `json:"type"`
	Text     string          `json:"text"`
	Thinking string          `json:"thinking"`
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
}

type usage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheReadTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_input_tokens"`
}

type userInput struct {
	Type    string      `json:"type"`
	Message userMessage `json:"message"`
}

type userMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// replyBuilder assembles one Messages API reply: its usage and stop reason
// from the stream events, its blocks from the assistant events.
type replyBuilder struct {
	started, stopped bool
	want             int // content blocks the stream announced
	blocks           []block
	in               usage
	out              int64
	stopReason       string
	uses             []*toolUse
}

func (r *replyBuilder) stream(e streamEvent) {
	switch e.Type {
	case streamMessageStart:
		*r = replyBuilder{started: true, in: e.Message.Usage}
	case streamBlockStart:
		r.want++
	case streamMessageDelta:
		r.stopReason, r.out = e.Delta.StopReason, e.Usage.OutputTokens
	case streamMessageStop:
		r.stopped = true
	}
}

func (r *replyBuilder) assistant(m apiMessage) {
	r.blocks = append(r.blocks, m.Content...)
}

func (r *replyBuilder) done() bool { return r.started && r.stopped && len(r.blocks) >= r.want }

func (r *replyBuilder) reply() inference.Reply {
	msg := inference.Message{Role: inference.RoleAssistant}
	var text, thinking []string
	for _, b := range r.blocks {
		switch b.Type {
		case blockText:
			text = append(text, b.Text)
		case blockThinking:
			if b.Thinking != "" {
				thinking = append(thinking, b.Thinking)
			}
		case blockToolUse:
			name := strings.TrimPrefix(b.Name, mcpToolPrefix)
			tc := inference.ToolCall{ID: b.ID, Type: "function"}
			tc.Function.Name, tc.Function.Arguments = name, string(b.Input)
			msg.ToolCalls = append(msg.ToolCalls, tc)
			var args any
			json.Unmarshal(b.Input, &args)
			r.uses = append(r.uses, &toolUse{id: b.ID, name: name, args: args})
		}
	}
	msg.Content, msg.Reasoning = strings.Join(text, "\n"), strings.Join(thinking, "\n")
	return inference.Reply{
		Message:      msg,
		Truncated:    r.stopReason == stopMaxTokens,
		InputTokens:  r.in.InputTokens + r.in.CacheReadTokens + r.in.CacheCreationTokens, // input_tokens excludes cached ones
		OutputTokens: r.out,
		CachedTokens: r.in.CacheReadTokens,
	}
}

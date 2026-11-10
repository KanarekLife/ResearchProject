package claudecode

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mcbench/integrations/inference"
)

func parseEvents(t *testing.T, lines ...string) []event {
	t.Helper()
	out := make([]event, len(lines))
	for i, l := range lines {
		if err := json.Unmarshal([]byte(l), &out[i]); err != nil {
			t.Fatal(err)
		}
	}
	return out
}

// A reply is complete only once the stream has stopped and every block it
// announced has arrived, whichever comes first.
func TestReplyWaitsForEveryBlock(t *testing.T) {
	g := &game{}
	var r replyBuilder
	evs := parseEvents(t,
		`{"type":"stream_event","event":{"type":"message_start","message":{"usage":{"input_tokens":3,"cache_read_input_tokens":90,"cache_creation_input_tokens":7}}}}`,
		`{"type":"stream_event","event":{"type":"content_block_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start"}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Attack."}]}}`,
		`{"type":"stream_event","event":{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":12}}}`,
		`{"type":"stream_event","event":{"type":"message_stop"}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"mcp__mcbench__choose_option","input":{"option_id":2}}]}}`,
	)
	for i, e := range evs {
		if err := g.event(e, &r); err != nil {
			t.Fatal(err)
		}
		if r.done() != (i == len(evs)-1) {
			t.Fatalf("after event %d: done = %v", i, r.done())
		}
	}
	reply := r.reply()
	tc := reply.Message.ToolCalls
	if len(tc) != 1 || tc[0].ID != "t1" || tc[0].Function.Name != "choose_option" || tc[0].Function.Arguments != `{"option_id":2}` {
		t.Fatalf("tool calls = %+v", tc)
	}
	if reply.Message.Content != "Attack." || reply.InputTokens != 100 || reply.CachedTokens != 90 || reply.OutputTokens != 12 {
		t.Fatalf("reply = %+v", reply)
	}
}

// A tool call is answered with its own result whether the call or the result
// arrives first, and a call made before its reply is complete waits for it.
func TestCallsGetTheirOwnResults(t *testing.T) {
	g := &game{waiting: map[string]*call{}, results: map[string]string{}}
	newCall := func(args string) *call {
		var a any
		json.Unmarshal([]byte(args), &a)
		return &call{name: "get_card", args: a, out: make(chan string, 1)}
	}
	early := newCall(`{"name":"Rhino"}`)
	g.call(early) // before the reply that asks for it
	g.open = []*toolUse{{id: "a", name: "get_card", args: early.args}, {id: "b", name: "get_card", args: newCall(`{"name":"Shocker"}`).args}}
	for _, c := range g.early {
		g.call(c)
	}
	g.early = nil
	g.deliver([]inference.Message{{Role: inference.RoleTool, ToolCallID: "b", Content: "shocker"}})
	late := newCall(`{"name":"Shocker"}`)
	g.call(late) // after its result
	g.deliver([]inference.Message{{Role: inference.RoleTool, ToolCallID: "b", Content: "shocker"}, {Role: inference.RoleTool, ToolCallID: "a", Content: "rhino"}})

	if got := <-early.out; got != "rhino" {
		t.Fatalf("early call got %q", got)
	}
	if got := <-late.out; got != "shocker" {
		t.Fatalf("late call got %q", got)
	}
}

// A request a usage limit rejects is waited out, but never past the game's
// total rate-limit wait.
func TestUsageLimitIsWaitedOutWithinTheCap(t *testing.T) {
	g := &game{}
	var r replyBuilder
	evs := parseEvents(t,
		`{"type":"rate_limit_event","rate_limit_info":{"status":"rejected","rateLimitType":"five_hour","resetsAt":4102444800}}`,
		`{"type":"result","subtype":"success","is_error":true,"result":"limit reached"}`,
	)
	if err := g.event(evs[0], &r); err != nil {
		t.Fatal(err)
	}
	err := g.event(evs[1], &r)
	var limit *rateLimit
	if !errors.As(err, &limit) || limit.kind != "five_hour" {
		t.Fatalf("got %v, want the five_hour usage limit", err)
	}
	if err := g.waitOut(context.Background(), limit, 6*time.Hour); err == nil {
		t.Fatal("a reset years away was waited for despite the 6h cap")
	}
}

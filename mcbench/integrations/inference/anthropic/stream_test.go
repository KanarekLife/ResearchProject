package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mcbench/integrations/inference"
)

// A streamed message is rebuilt like a plain one: thinking with its signature,
// text, and a tool call whose input arrives in pieces; the content goes back
// verbatim as Native.
func TestStreamRebuildsTheMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ev := range []string{
			`{"type":"message_start","message":{"usage":{"input_tokens":5,"cache_read_input_tokens":20,"output_tokens":1}}}`,
			`{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"hm"}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"m"}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"SIG"}}`,
			`{"type":"content_block_stop","index":0}`,
			`{"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}`,
			`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"ok"}}`,
			`{"type":"content_block_start","index":2,"content_block":{"type":"tool_use","id":"t1","name":"choose_option","input":{}}}`,
			`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"{\"option"}}`,
			`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"_id\":4}"}}`,
			`{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":12}}`,
			`{"type":"message_stop"}`,
		} {
			fmt.Fprintf(w, "data: %s\n\n", ev)
		}
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 100, Stream: true})
	r, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := r.Message
	if m.Reasoning != "hmm" || m.Content != "ok" || len(m.ToolCalls) != 1 || m.ToolCalls[0].ID != "t1" || m.ToolCalls[0].Function.Arguments != `{"option_id":4}` {
		t.Errorf("message %+v", m)
	}
	if !strings.Contains(string(m.Native), `"signature":"SIG"`) {
		t.Errorf("signature is not in the native content: %s", m.Native)
	}
	if r.InputTokens != 25 || r.OutputTokens != 12 || r.CachedTokens != 20 {
		t.Errorf("usage %+v", r)
	}
}

func TestStreamCutOffIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data: {\"type\":\"message_start\",\"message\":{\"usage\":{}}}\n\n")
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 100, Stream: true})
	if _, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil); err == nil {
		t.Error("a cut stream decoded without error")
	}
}

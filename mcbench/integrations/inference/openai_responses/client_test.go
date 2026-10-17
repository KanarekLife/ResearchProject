package openairesponses

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcbench/integrations/inference"
)

const reply = `{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[
{"type":"reasoning","id":"rs_1","summary":[{"type":"summary_text","text":"hm"}]},
{"type":"message","content":[{"type":"output_text","text":"ok"}]},
{"type":"function_call","id":"fc_1","call_id":"call_9","name":"choose_option","arguments":"{\"option_id\":2}"}],
"usage":{"input_tokens":100,"output_tokens":30,"input_tokens_details":{"cached_tokens":64},"output_tokens_details":{"reasoning_tokens":20}}}`

func TestCompleteMapsRequestAndReply(t *testing.T) {
	var path, auth string
	var req map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, auth = r.URL.Path, r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &req)
		w.Write([]byte(reply))
	}))
	defer srv.Close()

	var call inference.ToolCall
	call.ID, call.Function.Name, call.Function.Arguments = "call_1", "get_state", ""
	msgs := []inference.Message{
		{Role: inference.RoleSystem, Content: "sys"},
		{Role: inference.RoleUser, Content: "go"},
		{Role: inference.RoleAssistant, ToolCalls: []inference.ToolCall{call}},
		{Role: inference.RoleTool, ToolCallID: "call_1", Content: `{"round":1}`},
	}
	tools := []inference.Tool{{Name: "get_state", Description: "d", Parameters: map[string]any{"type": "object"}}}
	r, err := New(Config{BaseURL: srv.URL, APIKey: "k", Model: "m", MaxTokens: 50, Temperature: -1}).Complete(context.Background(), msgs, tools)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/responses" || auth != "Bearer k" || req["instructions"] != "sys" || req["store"] != false {
		t.Errorf("path %q auth %q body %v", path, auth, req)
	}
	items := req["input"].([]any)
	if len(items) != 3 || items[1].(map[string]any)["type"] != "function_call" || items[2].(map[string]any)["type"] != "function_call_output" {
		t.Errorf("input = %v", items)
	}
	if _, has := items[1].(map[string]any)["id"]; has {
		t.Errorf("function_call item carries an id: %v", items[1])
	}
	if len(r.Message.ToolCalls) != 1 || r.Message.ToolCalls[0].ID != "call_9" || r.Message.Content != "ok" || r.Message.Reasoning != "hm" {
		t.Errorf("reply = %+v", r.Message)
	}
	if !r.Truncated || r.InputTokens != 100 || r.OutputTokens != 30 || r.CachedTokens != 64 || r.ReasoningTokens != 20 {
		t.Errorf("usage = %+v", r)
	}
}

// A decoded reply is sent back as its own output items, so the encrypted
// reasoning stays with the function call it led to.
func TestReplyIsSentBackVerbatim(t *testing.T) {
	var req map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		req = nil
		json.Unmarshal(b, &req)
		w.Write([]byte(reply))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 50, Temperature: -1})
	msgs := []inference.Message{{Role: inference.RoleUser, Content: "go"}}
	r, err := c.Complete(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if inc, _ := json.Marshal(req["include"]); string(inc) != `["reasoning.encrypted_content"]` {
		t.Errorf("include = %s", inc)
	}
	msgs = append(msgs, r.Message, inference.Message{Role: inference.RoleTool, ToolCallID: "call_9", Content: "{}"})
	if _, err := c.Complete(context.Background(), msgs, nil); err != nil {
		t.Fatal(err)
	}
	var orig struct {
		Output []any `json:"output"`
	}
	json.Unmarshal([]byte(reply), &orig)
	got, _ := json.Marshal(req["input"].([]any)[1:4])
	if want, _ := json.Marshal(orig.Output); string(got) != string(want) {
		t.Errorf("reply sent back as %s, want %s", got, want)
	}
}

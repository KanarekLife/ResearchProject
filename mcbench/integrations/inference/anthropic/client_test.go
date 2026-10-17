package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcbench/integrations/inference"
)

const toolReply = `{"content":[{"type":"text","text":"ok"},{"type":"tool_use","id":"tu_2","name":"choose_option","input":{"option_id":3}}],
"stop_reason":"max_tokens","usage":{"input_tokens":10,"cache_read_input_tokens":80,"cache_creation_input_tokens":5,"output_tokens":7}}`

func TestCompleteMapsRequestAndReply(t *testing.T) {
	var path, key, ver string
	var req map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, key, ver = r.URL.Path, r.Header.Get("x-api-key"), r.Header.Get("anthropic-version")
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &req)
		w.Write([]byte(toolReply))
	}))
	defer srv.Close()

	var call inference.ToolCall
	call.ID, call.Function.Name, call.Function.Arguments = "tu_1", "get_state", ""
	msgs := []inference.Message{
		{Role: inference.RoleSystem, Content: "sys"},
		{Role: inference.RoleUser, Content: "go"},
		{Role: inference.RoleAssistant, ToolCalls: []inference.ToolCall{call}},
		{Role: inference.RoleTool, ToolCallID: "tu_1", Content: `{"round":1}`},
		{Role: inference.RoleUser, Content: "again"},
	}
	tools := []inference.Tool{{Name: "get_state", Description: "d", Parameters: map[string]any{"type": "object"}}}
	r, err := New(Config{BaseURL: srv.URL, APIKey: "k", Model: "claude", MaxTokens: 50, Temperature: -1}).Complete(context.Background(), msgs, tools)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/messages" || key != "k" || ver == "" || req["system"] != "sys" {
		t.Errorf("path %q key %q version %q system %v", path, key, ver, req["system"])
	}
	turns := req["messages"].([]any)
	if len(turns) != 3 {
		t.Fatalf("turns = %d, want 3", len(turns))
	}
	last := turns[2].(map[string]any)["content"].([]any)
	if len(last) != 2 || last[0].(map[string]any)["type"] != "tool_result" || last[1].(map[string]any)["type"] != "text" {
		t.Errorf("last turn = %v, want tool_result then text", last)
	}
	if len(r.Message.ToolCalls) != 1 || r.Message.ToolCalls[0].ID != "tu_2" || r.Message.ToolCalls[0].Function.Arguments != `{"option_id":3}` {
		t.Errorf("tool calls = %+v", r.Message.ToolCalls)
	}
	if !r.Truncated || r.InputTokens != 95 || r.CachedTokens != 80 || r.OutputTokens != 7 || r.Message.Content != "ok" {
		t.Errorf("reply = %+v", r)
	}
}

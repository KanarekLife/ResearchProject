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
	if cc, _ := req["cache_control"].(map[string]any); cc["type"] != "ephemeral" {
		t.Errorf("cache_control = %v, want the prefix cached", req["cache_control"])
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

// A reply with thinking goes back exactly as it came, signature and block
// order included, and its thinking is kept as the message's reasoning.
func TestThinkingIsSentBackVerbatim(t *testing.T) {
	const content = `[{"type":"thinking","thinking":"attack","signature":"sig=="},{"type":"tool_use","id":"tu_1","name":"choose_option","input":{"option_id":2}}]`
	var reqs []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &req)
		reqs = append(reqs, req)
		w.Write([]byte(`{"content":` + content + `,"stop_reason":"tool_use","usage":{}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Model: "claude", MaxTokens: 50, Temperature: -1})
	msgs := []inference.Message{{Role: inference.RoleUser, Content: "go"}}
	r, err := c.Complete(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Message.Reasoning != "attack" || len(r.Message.ToolCalls) != 1 {
		t.Fatalf("reply = %+v", r.Message)
	}
	msgs = append(msgs, r.Message, inference.Message{Role: inference.RoleTool, ToolCallID: "tu_1", Content: "{}"})
	if _, err := c.Complete(context.Background(), msgs, nil); err != nil {
		t.Fatal(err)
	}
	sent, _ := json.Marshal(reqs[1]["messages"].([]any)[1].(map[string]any)["content"])
	var want any // both sides re-encoded, so key order matches
	json.Unmarshal([]byte(content), &want)
	if wb, _ := json.Marshal(want); string(wb) != string(sent) {
		t.Errorf("assistant content sent back = %s, want %s", sent, wb)
	}
}

// An effort level is sent with adaptive thinking and wins over a budget; a
// budget alone is sent as fixed-budget thinking.
func TestReasoningSettings(t *testing.T) {
	for _, tc := range []struct {
		effort         string
		budget         int
		thinking, conf string
	}{
		{"", 0, `null`, `null`},
		{"high", 2048, `{"display":"summarized","type":"adaptive"}`, `{"effort":"high"}`},
		{"", 4096, `{"budget_tokens":4096,"type":"enabled"}`, `null`},
	} {
		b := New(Config{Model: "claude", MaxTokens: 50, Temperature: -1, ReasoningEffort: tc.effort, ReasoningBudget: tc.budget}).body(nil, nil)
		think, _ := json.Marshal(b["thinking"])
		conf, _ := json.Marshal(b["output_config"])
		if string(think) != tc.thinking || string(conf) != tc.conf {
			t.Errorf("effort %q budget %d: thinking %s output_config %s", tc.effort, tc.budget, think, conf)
		}
	}
}

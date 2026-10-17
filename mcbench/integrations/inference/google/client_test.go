package google

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mcbench/integrations/inference"
)

const toolReply = `{"candidates":[{"content":{"role":"model","parts":[
{"text":"thinking","thought":true},
{"functionCall":{"name":"choose_option","args":{"option_id":2}},"thoughtSignature":"SIG"}]},"finishReason":"STOP"}],
"usageMetadata":{"promptTokenCount":100,"candidatesTokenCount":10,"thoughtsTokenCount":40,"cachedContentTokenCount":60}}`

func TestCompleteMapsRequestAndReply(t *testing.T) {
	var path, key string
	var req map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, key = r.URL.Path, r.Header.Get("x-goog-api-key")
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &req)
		w.Write([]byte(toolReply))
	}))
	defer srv.Close()

	tools := []inference.Tool{
		{Name: "get_state", Description: "d", Parameters: map[string]any{"type": "object", "properties": map[string]any{}, "required": []string{}, "additionalProperties": false}},
		{Name: "choose_option", Description: "d", Parameters: map[string]any{"type": "object", "properties": map[string]any{"option_id": map[string]any{"type": "integer"}}, "required": []string{"option_id"}, "additionalProperties": false}},
	}
	var call inference.ToolCall
	call.ID, call.Function.Name, call.Function.Arguments = "get_state|S0", "get_state", ""
	msgs := []inference.Message{
		{Role: inference.RoleSystem, Content: "sys"},
		{Role: inference.RoleUser, Content: "go"},
		{Role: inference.RoleAssistant, ToolCalls: []inference.ToolCall{call}},
		{Role: inference.RoleTool, ToolCallID: call.ID, Content: `{"round":1}`},
		{Role: inference.RoleUser, Content: "again"},
	}
	r, err := New(Config{BaseURL: srv.URL, APIKey: "k", Model: "gem", MaxTokens: 50, Temperature: -1}).Complete(context.Background(), msgs, tools)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/models/gem:generateContent" || key != "k" {
		t.Errorf("path %q key %q", path, key)
	}
	contents := req["contents"].([]any)
	if len(contents) != 3 { // user, model, then the response and the text share one user content
		t.Fatalf("contents = %d, want 3: %v", len(contents), contents)
	}
	model := contents[1].(map[string]any)["parts"].([]any)[0].(map[string]any)
	if model["thoughtSignature"] != "S0" || model["functionCall"].(map[string]any)["name"] != "get_state" {
		t.Errorf("model part = %v", model)
	}
	if n := len(contents[2].(map[string]any)["parts"].([]any)); n != 2 {
		t.Errorf("last user content has %d parts, want response + text", n)
	}
	decls := req["tools"].([]any)[0].(map[string]any)["functionDeclarations"].([]any)
	if _, has := decls[0].(map[string]any)["parameters"]; has {
		t.Errorf("tool without properties kept a parameters schema")
	}
	if p := decls[1].(map[string]any)["parameters"].(map[string]any); p["additionalProperties"] != nil {
		t.Errorf("additionalProperties sent: %v", p)
	}

	if len(r.Message.ToolCalls) != 1 || r.Message.ToolCalls[0].ID != "choose_option|SIG" || r.Message.ToolCalls[0].Function.Arguments != `{"option_id":2}` {
		t.Errorf("tool calls = %+v", r.Message.ToolCalls)
	}
	if r.Message.Reasoning != "thinking" || r.InputTokens != 100 || r.OutputTokens != 50 || r.CachedTokens != 60 || r.ReasoningTokens != 40 {
		t.Errorf("reply = %+v", r)
	}
}

func TestCompleteRetriesRateLimit(t *testing.T) {
	inference.BackoffBase = time.Millisecond
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls++; calls < 2 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(toolReply))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 1, Retries: 2, Temperature: -1})
	if _, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil); err != nil || calls != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestThinkingLevelWinsOverBudget(t *testing.T) {
	body := New(Config{Model: "m", MaxTokens: 1, Temperature: -1, ReasoningEffort: "medium", ReasoningBudget: 2048}).body(nil, nil)
	think := body["generationConfig"].(map[string]any)["thinkingConfig"].(map[string]any)
	if think["thinkingLevel"] != "medium" || think["thinkingBudget"] != nil {
		t.Errorf("thinkingConfig = %v, want only the level", think)
	}
	body = New(Config{Model: "m", MaxTokens: 1, Temperature: -1, ReasoningBudget: 2048}).body(nil, nil)
	if think = body["generationConfig"].(map[string]any)["thinkingConfig"].(map[string]any); think["thinkingBudget"] != 2048 {
		t.Errorf("thinkingConfig = %v, want the budget", think)
	}
}

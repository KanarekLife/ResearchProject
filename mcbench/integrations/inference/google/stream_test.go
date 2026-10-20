package google

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mcbench/integrations/inference"
)

// A streamed reply is its chunks joined: text split mid-word is one text,
// thoughts stay apart from the answer, and the last usage and finish reason win.
func TestStreamJoinsTheChunks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ":streamGenerateContent") || r.URL.Query().Get("alt") != "sse" {
			t.Errorf("request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		for _, ev := range []string{
			`{"candidates":[{"content":{"role":"model","parts":[{"text":"thi","thought":true}]}}]}`,
			`{"candidates":[{"content":{"role":"model","parts":[{"text":"nk","thought":true},{"text":"an"}]}}]}`,
			`{"candidates":[{"content":{"role":"model","parts":[{"text":"swer"},{"functionCall":{"name":"choose_option","args":{"option_id":2}},"thoughtSignature":"S"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":8,"candidatesTokenCount":3,"thoughtsTokenCount":2}}`,
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
	if m.Reasoning != "think" || m.Content != "answer" || len(m.ToolCalls) != 1 || m.ToolCalls[0].ID != "choose_option|S" || m.ToolCalls[0].Function.Arguments != `{"option_id":2}` {
		t.Errorf("message %+v", m)
	}
	if r.InputTokens != 8 || r.OutputTokens != 5 || r.ReasoningTokens != 2 {
		t.Errorf("usage %+v", r)
	}
}

func TestStreamCutOffIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"par\"}]}}]}\n\n")
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 100, Stream: true})
	if _, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil); err == nil {
		t.Error("a cut stream decoded without error")
	}
}

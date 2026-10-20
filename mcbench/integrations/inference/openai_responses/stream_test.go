package openairesponses

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcbench/integrations/inference"
)

// A streamed reply is the response in its last event.
func TestStreamTakesTheCompletedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"he\"}\n\n")
		fmt.Fprint(w, "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"output\":[{\"type\":\"function_call\",\"call_id\":\"c1\",\"name\":\"choose_option\",\"arguments\":\"{\\\"option_id\\\":3}\"}],\"usage\":{\"input_tokens\":9,\"output_tokens\":4}}}\n\n")
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 10, Stream: true})
	r, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Message.ToolCalls) != 1 || r.Message.ToolCalls[0].Function.Arguments != `{"option_id":3}` || r.InputTokens != 9 || r.OutputTokens != 4 {
		t.Errorf("reply %+v", r)
	}
}

func TestStreamWithoutACompletedResponseIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"he\"}\n\n")
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, Model: "m", MaxTokens: 10, Stream: true})
	if _, err := c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil); err == nil {
		t.Error("a cut stream decoded without error")
	}
}

package openaicompatible

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mcbench/integrations/inference"
)

const okBody = `{"choices":[{"message":{"role":"assistant","content":"hi"},"finish_reason":"%s"}],
"usage":{"prompt_tokens":10,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":4}}}`

// serve runs handler as the endpoint and returns a client for it.
func serve(t *testing.T, retries int, handler http.HandlerFunc) *Client {
	t.Helper()
	backoffBase = time.Millisecond
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL, APIKey: "k", Model: "m", MaxTokens: 100, Retries: retries})
}

func complete(c *Client) (inference.Reply, error) {
	return c.Complete(context.Background(), []inference.Message{{Role: inference.RoleUser, Content: "x"}}, nil)
}

func TestCompleteDecodesReplyAndUsage(t *testing.T) {
	c := serve(t, 0, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		fmt.Fprintf(w, okBody, "length")
	})
	r, err := complete(c)
	if err != nil {
		t.Fatal(err)
	}
	if r.Message.Content != "hi" || !r.Truncated || r.InputTokens != 10 || r.OutputTokens != 5 || r.CachedTokens != 4 {
		t.Errorf("reply = %+v", r)
	}
}

func TestCompleteRetriesTransientFailures(t *testing.T) {
	calls := 0
	c := serve(t, 2, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintf(w, okBody, "stop")
	})
	if _, err := complete(c); err != nil || calls != 3 {
		t.Fatalf("err=%v after %d calls, want success on call 3", err, calls)
	}
}

func TestCompleteDoesNotRetryClientErrors(t *testing.T) {
	calls := 0
	c := serve(t, 3, func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "bad request", http.StatusUnauthorized)
	})
	if _, err := complete(c); err == nil || calls != 1 {
		t.Fatalf("err=%v after %d calls, want an error after 1", err, calls)
	}
}

func TestCompleteGivesUpAfterRetries(t *testing.T) {
	calls := 0
	c := serve(t, 1, func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "busy", http.StatusBadGateway)
	})
	_, err := complete(c)
	if err == nil || calls != 2 || !strings.Contains(err.Error(), "2 attempts") {
		t.Fatalf("err=%v after %d calls, want failure after 2 attempts", err, calls)
	}
}

// Some servers only accept max_completion_tokens; the client switches once,
// without spending a retry.
func TestCompleteSwitchesTokenParameter(t *testing.T) {
	var fields []string
	c := serve(t, 0, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body[fieldMaxCompletionTokens]; ok {
			fields = append(fields, fieldMaxCompletionTokens)
			fmt.Fprintf(w, okBody, "stop")
			return
		}
		fields = append(fields, fieldMaxTokens)
		http.Error(w, `unsupported parameter: use max_completion_tokens`, http.StatusBadRequest)
	})
	if _, err := complete(c); err != nil {
		t.Fatal(err)
	}
	if strings.Join(fields, ",") != fieldMaxTokens+","+fieldMaxCompletionTokens {
		t.Errorf("token fields tried: %v", fields)
	}
}

func TestCompleteRejectsEmptyChoices(t *testing.T) {
	c := serve(t, 0, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"choices":[]}`) })
	if _, err := complete(c); err == nil {
		t.Fatal("a reply without choices was accepted")
	}
}

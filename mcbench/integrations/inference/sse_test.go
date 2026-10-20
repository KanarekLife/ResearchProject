package inference

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostSSEDeliversEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, ": keep-alive\n\nevent: a\ndata: {\"n\":1}\n\ndata: {\"n\":\ndata: 2}\n\n")
	}))
	defer srv.Close()
	type ev struct{ name, data string }
	var got []ev
	err := PostSSE(context.Background(), srv.URL, nil, []byte(`{}`), func(name string, data []byte) error {
		got = append(got, ev{name, string(data)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != (ev{"a", "{\"n\":1}\n"}) || got[1].name != "" || got[1].data != "{\"n\":\n2}\n" {
		t.Errorf("events %q", got)
	}
}

func TestPostSSEReportsHTTPErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	defer srv.Close()
	err := PostSSE(context.Background(), srv.URL, nil, []byte(`{}`), func(string, []byte) error { return nil })
	var he *HTTPError
	if !errors.As(err, &he) || he.Status != http.StatusTooManyRequests {
		t.Errorf("error %v, want an HTTP 429", err)
	}
}

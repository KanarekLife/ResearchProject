package systemone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChoose(t *testing.T) {
	var got request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" || r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("request %s with auth %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"model":"m","answers":{"decision":{"type":"choice","choice":"2","confidence":0.9,"probabilities":{"1":0.1,"2":0.9}}},"usage":{"input_tokens":7}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"})
	ans, err := c.Choose(context.Background(), map[string]int{"hp": 3}, "pick", []Option{{"1", "attack"}, {"2", "end turn"}})
	if err != nil {
		t.Fatal(err)
	}
	if ans.Choice != "2" || ans.Probabilities["2"] != 0.9 || ans.InputTokens != 7 {
		t.Errorf("answer %+v", ans)
	}
	q := got.Questions[questionName]
	if got.Model != "m" || q.Type != questionTypeChoice || q.Criteria["1"] != "attack" || q.Criteria["2"] != "end turn" {
		t.Errorf("request %+v", got)
	}
}

func TestChooseWithoutAnswer(t *testing.T) {
	if _, err := decode([]byte(`{"answers":{}}`)); err == nil {
		t.Error("a reply with no decision answer decoded without error")
	}
}

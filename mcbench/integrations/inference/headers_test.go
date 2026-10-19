package inference

import (
	"context"
	"testing"
)

func TestExpandHeaders(t *testing.T) {
	h := map[string]string{"x-opencode-session": "{run}/{game}", "x-plain": "v"}
	got := ExpandHeaders(WithGame(context.Background(), "r1", "g2"), h)
	if got["x-opencode-session"] != "r1/g2" || got["x-plain"] != "v" {
		t.Fatalf("got %v", got)
	}
	if h["x-opencode-session"] != "{run}/{game}" {
		t.Fatal("input headers were modified")
	}
}

func TestValidateHeaderValue(t *testing.T) {
	if err := ValidateHeaderValue("a-{run}-{game}"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHeaderValue("{seed}"); err == nil {
		t.Fatal("unknown placeholder accepted")
	}
}

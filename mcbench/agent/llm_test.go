package agent

import (
	"encoding/json"
	"testing"
)

func TestParseAnswer(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{`{"reasoning": "lethal", "option_id": 7}`, 7},
		{"```json\n{\"reasoning\": \"x\", \"option_id\": 3}\n```", 3},
		{`<think>maybe {"option_id": 1}</think> Final: {"reasoning": "y", "option_id": 12}`, 12},
		{`I pick option_id: 4`, 4},
		{`I pick the attack.`, 0},
	}
	for _, c := range cases {
		raw, _ := parseAnswer(c.in)
		var a struct {
			OptionID int `json:"option_id"`
		}
		json.Unmarshal(raw, &a)
		if a.OptionID != c.want {
			t.Errorf("parseAnswer(%q) = %d, want %d", c.in, a.OptionID, c.want)
		}
	}
}

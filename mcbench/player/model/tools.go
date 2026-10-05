package model

import (
	"encoding/json"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/integrations/inference"
)

// toolSchemas converts the game's tool contract into the inference client's
// tool type. The game owns the schemas; this package only forwards them.
func toolSchemas() []inference.Tool {
	out := make([]inference.Tool, len(session.Tools))
	for i, t := range session.Tools {
		out[i] = inference.Tool{Name: t.Name, Description: t.Description, Parameters: t.Parameters}
	}
	return out
}

// toolResult is the outcome of one assistant reply's tool calls.
type toolResult struct {
	messages []inference.Message // tool result messages to append
	events   []string            // events from a successful choose_option
	chosen   bool                // a choose_option succeeded
}

// runTools executes the tool calls in order. Only the first choose_option is
// applied; later ones get an error result.
func runTools(s *session.Session, calls []inference.ToolCall) toolResult {
	var res toolResult
	for _, tc := range calls {
		out := `{"error":"the decision was already made in this reply"}`
		if !res.chosen {
			out = s.Call(tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			if tc.Function.Name == constants.ToolChooseOption {
				res.events, res.chosen = chosen(out)
			}
		}
		res.messages = append(res.messages, inference.Message{Role: inference.RoleTool, ToolCallID: tc.ID, Content: out})
	}
	return res
}

// chosen reports whether a choose_option result succeeded, with its events.
func chosen(result string) ([]string, bool) {
	var r struct {
		Error  string   `json:"error"`
		Events []string `json:"events"`
	}
	if json.Unmarshal([]byte(result), &r) != nil || r.Error != "" {
		return nil, false
	}
	return r.Events, true
}

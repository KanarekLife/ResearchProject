package model

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/integrations/inference"
	"mcbench/player"
)

// traceLine is a message or choice line of a game trace, whichever fields the
// line has.
type traceLine struct {
	Type         string            `json:"type"`
	Decision     int               `json:"decision"`
	Message      inference.Message `json:"message"`
	InputTokens  int64             `json:"prompt_tokens"`
	OutputTokens int64             `json:"completion_tokens"`
	CachedTokens int64             `json:"cached_tokens"`
	Truncated    bool              `json:"truncated"`
	Choice       session.Choice    `json:"choice"`
}

// restore rebuilds a game from its trace: the conversation up to its last
// complete exchange, and the board by replaying the choices made in it. Choices
// and messages go back into the session's trace in their original order. It
// returns nil messages when the trace holds no conversation to continue.
func (m *Model) restore(s *session.Session, trace []byte, u *player.Usage) ([]inference.Message, error) {
	var messages, choices []traceLine
	sc := bufio.NewScanner(bytes.NewReader(trace))
	sc.Buffer(nil, 1<<26)
	for sc.Scan() {
		var l traceLine
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue // a line cut off when the run was killed
		}
		switch l.Type {
		case constants.TraceMessage:
			messages = append(messages, l)
		case constants.TraceChoice:
			choices = append(choices, l)
		}
	}
	messages = messages[:completePrefix(messages)]
	if len(messages) < 2 {
		return nil, nil
	}

	var msgs []inference.Message
	next := 0
	for _, l := range messages {
		// A message of decision d comes after the choices that led to d.
		for ; next < len(choices) && choices[next].Decision < l.Decision; next++ {
			if err := replayChoice(s, choices[next].Choice); err != nil {
				return nil, fmt.Errorf("replaying decision %d: %w", choices[next].Decision, err)
			}
		}
		reply := inference.Reply{InputTokens: l.InputTokens, OutputTokens: l.OutputTokens, CachedTokens: l.CachedTokens, Truncated: l.Truncated}
		msgs = m.add(s, msgs, l.Message, &reply)
		if l.Message.Role == inference.RoleAssistant {
			u.Requests++
			u.InputTokens += l.InputTokens
			u.OutputTokens += l.OutputTokens
			u.CachedTokens += l.CachedTokens
		}
	}
	if want := messages[len(messages)-1].Decision - 1; len(s.Choices) != want {
		return nil, fmt.Errorf("replay made %d choices, the conversation has %d", len(s.Choices), want)
	}
	return msgs, nil
}

// completePrefix is how many leading messages end on a user or tool message
// with every tool call answered, so the conversation can be sent as it is.
func completePrefix(ms []traceLine) int {
	pending := map[string]bool{}
	n := 0
	for i, l := range ms {
		switch l.Message.Role {
		case inference.RoleAssistant:
			if len(pending) > 0 {
				return n
			}
			for _, tc := range l.Message.ToolCalls {
				pending[tc.ID] = true
			}
		case inference.RoleTool:
			delete(pending, l.Message.ToolCallID)
		}
		if len(pending) == 0 && l.Message.Role != inference.RoleAssistant {
			n = i + 1
		}
	}
	return n
}

// replayChoice repeats a recorded choice. A forced choice is applied by the
// session itself, so it is skipped.
func replayChoice(s *session.Session, c session.Choice) error {
	if c.Auto {
		return nil
	}
	d := s.Decision()
	if d == nil {
		return fmt.Errorf("the game is over, but %q is recorded", c.Key)
	}
	for _, o := range d.Options {
		if o.Key == c.Key {
			_, err := s.Choose(o.ID, c.Reasoning)
			return err
		}
	}
	return fmt.Errorf("%q is not offered", c.Key)
}

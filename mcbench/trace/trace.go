// Package trace reads the JSON lines a game writes as it runs, so a viewer can
// follow a game from its start entry to its end entry.
package trace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Entry is one line of a game's trace. Every type writes into the same struct,
// so a reader decodes once and switches on Type.
type Entry struct {
	Type string `json:"type"`
	Time string `json:"time"`

	// start, end
	Scenario     string                   `json:"scenario"`
	Seed         uint64                   `json:"seed"`
	Sample       int                      `json:"sample"`
	Player       string                   `json:"player"`
	Instructions instruction.Instructions `json:"instructions"`
	Status       string                   `json:"status"`
	Error        string                   `json:"error"`
	Rounds       int                      `json:"rounds"`
	Criteria     map[string]float64       `json:"criteria"`
	Score        float64                  `json:"score"`
	Usage        player.Usage             `json:"usage"`

	// message
	Decision         int     `json:"decision"`
	Message          Message `json:"message"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	CachedTokens     int64   `json:"cached_tokens"`
	Truncated        bool    `json:"truncated"`

	// choice
	Choice session.Choice `json:"choice"`
	Text   string         `json:"text"`
	Events []string       `json:"events"`
}

// Message is the conversation message of a model reply. Only the fields a
// reader needs are decoded.
type Message struct {
	Role      string     `json:"role"`
	Reasoning string     `json:"reasoning_content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

// ToolCall is one function call the model requested.
type ToolCall struct {
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}

// Decode reads every trace line in r. A trailing partial line is ignored, so a
// trace being written as it is read stays readable.
func Decode(r io.Reader) ([]Entry, error) {
	var out []Entry
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<16), 1<<26)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 || line[len(line)-1] != '}' {
			continue // partial line from an in-progress write
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return out, fmt.Errorf("trace line %d: %w", len(out)+1, err)
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

// ReadFile decodes the trace at path.
func ReadFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

// Follower yields the trace entries of a growing file, remembering how far it
// has read. It is the live counterpart of Decode.
type Follower struct {
	f   *os.File
	rd  *bufio.Reader
	off int64
}

// Follow opens path for following. The file may still be being written.
func Follow(path string) (*Follower, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Follower{f: f, rd: bufio.NewReader(f)}, nil
}

// Next returns the entries appended since the last call, blocking up to poll for
// the first one. It returns nil when the file has not grown.
func (f *Follower) Next(poll time.Duration) ([]Entry, error) {
	if !waitFor(f.f, f.off, poll) {
		return nil, nil
	}
	if _, err := f.f.Seek(f.off, 0); err != nil {
		return nil, err
	}
	f.rd.Reset(f.f)
	var out []Entry
	for {
		line, err := f.rd.ReadBytes('\n')
		complete := err == nil && len(line) > 0 && line[len(line)-1] == '\n'
		if complete {
			f.off += int64(len(line))
			line = line[:len(line)-1]
			if len(line) == 0 {
				continue
			}
			var e Entry
			if err := json.Unmarshal(line, &e); err != nil {
				return out, fmt.Errorf("trace line %d: %w", len(out)+1, err)
			}
			out = append(out, e)
			continue
		}
		if err == io.EOF && len(line) > 0 {
			// partial trailing line: leave it for the next call
			if _, serr := f.f.Seek(f.off, 0); serr != nil {
				return out, serr
			}
			f.rd.Reset(f.f)
		}
		return out, errNotEOF(err)
	}
}

// Close releases the file.
func (f *Follower) Close() error { return f.f.Close() }

func errNotEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

// waitFor polls until the file is longer than off, or the timeout expires.
func waitFor(f *os.File, off int64, poll time.Duration) bool {
	deadline := time.Now().Add(poll)
	for {
		if fi, err := f.Stat(); err == nil && fi.Size() > off {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// IsChoice reports whether e is a decision the player made.
func IsChoice(e Entry) bool { return e.Type == constants.TraceChoice }

// IsEnd reports whether e closes the game.
func IsEnd(e Entry) bool { return e.Type == constants.TraceEnd }

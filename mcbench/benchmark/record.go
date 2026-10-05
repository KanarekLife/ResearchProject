// Package benchmark is the harness: it plays scenario games with a player,
// scores the final state and reports the results. It is deliberately small:
// Play runs one game, Run schedules many, Score and Report summarize them.
package benchmark

import (
	"bufio"
	"encoding/json"
	"os"

	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Record is one finished game. It holds everything needed to reproduce and
// inspect the result without replaying it.
type Record struct {
	RunID        string                   `json:"run_id"`
	Scenario     string                   `json:"scenario"`
	Player       string                   `json:"player"`
	Instructions instruction.Instructions `json:"instructions"`
	Seed         uint64                   `json:"seed"`
	Sample       int                      `json:"sample"`
	Status       string                   `json:"status"` // a session status, or "agent_error"
	Error        string                   `json:"error,omitempty"`
	Rounds       int                      `json:"rounds"`
	Decisions    int                      `json:"decisions"`
	Invalid      int                      `json:"invalid_choices,omitempty"`
	Stats        map[string]float64       `json:"stats"`
	Criteria     map[string]float64       `json:"criteria"`
	Score        float64                  `json:"score"`
	Usage        player.Usage             `json:"usage,omitzero"`
	Millis       int64                    `json:"ms"`
	Choices      []session.Choice         `json:"choices"`
	Log          []string                 `json:"log"`
}

// ReadRecords reads a games.jsonl file.
func ReadRecords(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<30)
	for sc.Scan() {
		var r Record
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, sc.Err()
}

package boardgui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"mcbench/constants"
	"mcbench/trace"
)

// stateLive is a game whose trace has no end entry yet.
const stateLive = "live"

// tailBytes is how much of a trace's end is read to find its end entry, which
// is short and always the last line.
const tailBytes = 16 << 10

// RunInfo is one run of the results directories as listed to the page.
type RunInfo struct {
	ID    string     `json:"id"`
	Games []GameItem `json:"games"`
	Local bool       `json:"local,omitempty"` // not published: it may still be played
}

// GameItem is one game of a run: its name and "live" or its end status.
type GameItem struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Steps string `json:"steps,omitempty"` // an exported site's replay file
}

// Runs is the page's picker: every run and the selection to open by default.
type Runs struct {
	Runs []RunInfo `json:"runs"`
	Run  string    `json:"run"`
	Game string    `json:"game"` // empty: follow the run's current game
	// Static is an exported site.
	Static bool `json:"static,omitempty"`
}

// peek reads when a game started (zero before its start entry is written) and
// whether it has ended.
func peek(path string) (started time.Time, state string) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, stateLive
	}
	defer f.Close()
	if line, err := bufio.NewReader(f).ReadBytes('\n'); err == nil {
		var e trace.Entry
		if json.Unmarshal(line, &e) == nil && e.Type == constants.TraceStart {
			started, _ = time.Parse(time.RFC3339, e.Time)
		}
	}
	state = stateLive
	if fi, err := f.Stat(); err == nil {
		f.Seek(max(fi.Size()-tailBytes, 0), io.SeekStart)
		tail, _ := io.ReadAll(f)
		lines := bytes.Split(bytes.TrimRight(tail, "\n"), []byte("\n"))
		var e trace.Entry
		if json.Unmarshal(lines[len(lines)-1], &e) == nil && e.Type == constants.TraceEnd {
			state = e.Status
		}
	}
	return started, state
}

// current picks the game a viewer following the run should see: the earliest
// started game still being played or, when every game has ended, the last one
// started. A game stays current until it ends, so with -parallel the view does
// not jump between games written at the same time, and a newly started game is
// picked up once the one before it ends.
func current(run trace.Run) trace.Game {
	var live, last trace.Game
	var liveAt, lastAt time.Time
	for _, g := range run.Games {
		at, state := peek(g.Trace)
		if at.IsZero() {
			continue // just created: its start entry is not written yet
		}
		if state == stateLive && (live.Trace == "" || at.Before(liveAt)) {
			live, liveAt = g, at
		}
		if last.Trace == "" || !at.Before(lastAt) {
			last, lastAt = g, at
		}
	}
	switch {
	case live.Trace != "":
		return live
	case last.Trace != "":
		return last
	}
	return run.Games[len(run.Games)-1]
}

// listRuns lists the runs under dirs with each game's state.
func listRuns(dirs []string, online string) ([]RunInfo, error) {
	runs, err := trace.Discover(dirs...)
	if err != nil {
		return nil, err
	}
	out := make([]RunInfo, len(runs))
	for i, r := range runs {
		out[i].ID = r.ID
		out[i].Local = online == "" || filepath.Dir(r.Dir) != filepath.Clean(online)
		for _, g := range r.Games {
			_, state := peek(g.Trace)
			out[i].Games = append(out[i].Games, GameItem{Name: g.Name, State: state})
		}
	}
	return out, nil
}

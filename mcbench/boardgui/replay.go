package boardgui

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/trace"
)

// A trace records each choice but not the board, so the viewer rebuilds the
// boards by replaying the recorded choices on the same scenario and seed. The
// game is deterministic, so the replay is the game that was played; each
// choice's events are compared with the recorded ones to prove it.

// Step is the board at one decision: what the player saw, the decision it
// faced and, once the trace has it, what it chose and what followed. The last
// step of a finished game has no decision.
type Step struct {
	Board    session.View      `json:"board"`
	Status   string            `json:"status"`
	Decision *session.Decision `json:"decision,omitempty"`
	Choice   *Chosen           `json:"choice,omitempty"`
}

// Chosen is the option the trace recorded for a decision.
type Chosen struct {
	Key       string `json:"key"`
	Text      string `json:"text"`
	Reasoning string `json:"reasoning,omitempty"`
	// Thinking is the model's reasoning_content from its replies to this
	// decision; scripted players have none.
	Thinking string   `json:"thinking,omitempty"`
	Events   []string `json:"events"`
}

// GameInfo is what the trace's start entry says about the game.
type GameInfo struct {
	Scenario string `json:"scenario"`
	Seed     uint64 `json:"seed"`
	Sample   int    `json:"sample"`
	Player   string `json:"player"`
}

// Result is what the trace's end entry says about the game.
type Result struct {
	Status   string             `json:"status"`
	Error    string             `json:"error,omitempty"`
	Rounds   int                `json:"rounds"`
	Score    float64            `json:"score"`
	Criteria map[string]float64 `json:"criteria,omitempty"`
}

// Snapshot is the replay as served to the page: the steps from From on.
type Snapshot struct {
	// Run and Game name the trace; a page following a run sees Game change
	// when the next game starts.
	Run   string    `json:"run"`
	Game  string    `json:"game"`
	Info  *GameInfo `json:"info,omitempty"`
	End   *Result   `json:"end,omitempty"`
	Error string    `json:"error,omitempty"`
	Total int       `json:"total"`
	// Rebuilds changes when a retry replaced earlier steps; a reader that
	// holds steps from before must fetch them again.
	Rebuilds int    `json:"rebuilds"`
	From     int    `json:"from"`
	Steps    []Step `json:"steps"`
}

// replay rebuilds a game's boards from its trace entries. It is safe for one
// writer (apply) and many readers (snapshot).
//
// A choice is recorded by its option key, and two copies of a card in hand
// share a key, so a recorded choice can match several options. The replay
// takes the first and, if the game later diverges from the trace, retries the
// other copies (latest first) until the whole trace replays.
type replay struct {
	scenarios []*scenario.Scenario

	mu       sync.Mutex
	entries  []trace.Entry
	picks    []int // per choice: which of its matching options was taken
	matches  []int // per choice: how many options matched its key
	err      string
	rebuilds int
	stopped  bool // the viewer dropped this replay

	// rebuilt from entries and picks
	game     *GameInfo
	end      *Result
	sess     *session.Session
	steps    []Step
	choices  int
	thinking []string // the current decision's model reasoning so far
}

func newReplay(scs []*scenario.Scenario) *replay { return &replay{scenarios: scs} }

// quiet discards the replayed game's action log; the game was logged when it
// was played.
var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

// maxRebuilds bounds the search for which copy of a card each ambiguous choice
// took.
const maxRebuilds = 1000

// apply feeds trace entries to the replay. After an error it cannot recover
// from, the replay stops and keeps the steps it has.
func (r *replay) apply(entries []trace.Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range entries {
		if r.err != "" {
			return
		}
		r.entries = append(r.entries, e)
		if err := r.applyOne(e); err != nil && !r.retry() {
			r.err = err.Error()
		}
	}
}

// retry replays the trace again with other copies taken at ambiguous choices,
// until one replays every entry so far.
func (r *replay) retry() bool {
	for range maxRebuilds {
		if !r.nextPicks() {
			return false
		}
		if r.rebuild() == nil {
			r.rebuilds++
			return true
		}
	}
	return false
}

// nextPicks moves to the next untried combination of picks up to the choice
// that failed, changing the latest ambiguous choice first.
func (r *replay) nextPicks() bool {
	for a := min(r.choices, len(r.picks)) - 1; a >= 0; a-- {
		if r.picks[a]+1 < r.matches[a] {
			r.picks[a]++
			r.picks, r.matches = r.picks[:a+1], r.matches[:a+1]
			return true
		}
	}
	return false
}

// rebuild replays every entry from the start with the current picks.
func (r *replay) rebuild() error {
	r.game, r.end, r.sess, r.steps, r.choices, r.thinking = nil, nil, nil, nil, 0, nil
	for _, e := range r.entries {
		if err := r.applyOne(e); err != nil {
			return err
		}
	}
	return nil
}

func (r *replay) applyOne(e trace.Entry) error {
	switch e.Type {
	case constants.TraceStart:
		return r.start(e)
	case constants.TraceMessage:
		if e.Message.Reasoning != "" {
			r.thinking = append(r.thinking, e.Message.Reasoning)
		}
	case constants.TraceChoice:
		return r.choose(e)
	case constants.TraceEnd:
		r.end = &Result{Status: e.Status, Error: e.Error, Rounds: e.Rounds, Score: e.Score, Criteria: e.Criteria}
	}
	return nil
}

// start sets up the recorded scenario and seed. Option order does not matter:
// choices are matched by their stable key, so the session is not shuffled.
func (r *replay) start(e trace.Entry) error {
	i := slices.IndexFunc(r.scenarios, func(s *scenario.Scenario) bool { return s.ID == e.Scenario })
	if i < 0 {
		return fmt.Errorf("scenario %q is not in the scenario directory", e.Scenario)
	}
	sc := r.scenarios[i]
	s, err := session.New(sc, e.Seed, session.Options{MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions, Logger: quiet})
	if err != nil {
		return err
	}
	r.game = &GameInfo{Scenario: e.Scenario, Seed: e.Seed, Sample: e.Sample, Player: e.Player}
	r.sess = s
	r.record()
	return nil
}

// choose replays one recorded choice and checks it produced the recorded events.
func (r *replay) choose(e trace.Entry) error {
	if r.sess == nil {
		return fmt.Errorf("decision %d comes before the start entry", e.Decision)
	}
	last := &r.steps[len(r.steps)-1]
	if last.Decision == nil {
		return fmt.Errorf("decision %d: the replayed game is already over (%s)", e.Decision, last.Status)
	}
	n := r.choices
	r.choices++
	ids := optionIDs(last.Decision, e.Choice.Key, e.Text)
	if len(ids) == 0 {
		return fmt.Errorf("decision %d: option %q is not offered in the replay (engine or data changed since the trace?)", e.Decision, e.Choice.Key)
	}
	if n == len(r.picks) {
		r.picks = append(r.picks, 0)
		r.matches = append(r.matches, 0)
	}
	r.matches[n] = len(ids)
	events, err := r.sess.Choose(ids[r.picks[n]], "")
	if err != nil {
		return fmt.Errorf("decision %d: %w", e.Decision, err)
	}
	last.Choice = &Chosen{Key: e.Choice.Key, Text: e.Text, Reasoning: e.Choice.Reasoning, Thinking: strings.Join(r.thinking, "\n\n"), Events: e.Events}
	r.thinking = nil
	if !slices.Equal(events, e.Events) {
		return fmt.Errorf("decision %d: the replay diverged from the trace (engine or data changed since the trace?)", e.Decision)
	}
	r.record()
	return nil
}

// optionIDs lists the decision's options with the recorded key and text, or
// with the key alone when none has the text.
func optionIDs(d *session.Decision, key, text string) []int {
	var exact, byKey []int
	for _, o := range d.Options {
		if o.Key != key {
			continue
		}
		byKey = append(byKey, o.ID)
		if o.Text == text {
			exact = append(exact, o.ID)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return byKey
}

// record appends the current board as a new step.
func (r *replay) record() {
	r.steps = append(r.steps, Step{Board: r.sess.View(), Status: r.sess.Status(), Decision: r.sess.Decision()})
}

// done reports whether nothing more can change: the game ended or the replay
// failed.
func (r *replay) done() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.end != nil || r.err != "" || r.stopped
}

// stop ends following the trace.
func (r *replay) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopped = true
}

// snapshot returns the steps from index from on.
func (r *replay) snapshot(from int) Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	from = min(max(from, 0), len(r.steps))
	return Snapshot{
		Info: r.game, End: r.end, Error: r.err, Total: len(r.steps), From: from, Rebuilds: r.rebuilds,
		Steps: append([]Step{}, r.steps[from:]...),
	}
}

// follow reads the trace at path into the replay as it grows, until the game
// ends or the replay fails. A finished trace is read in the first pass.
func (r *replay) follow(path string, poll time.Duration) {
	f, err := trace.Follow(path)
	if err != nil {
		r.fail(err)
		return
	}
	defer f.Close()
	for !r.done() {
		entries, err := f.Next(poll)
		r.apply(entries)
		if err != nil {
			r.fail(err)
			return
		}
	}
}

func (r *replay) fail(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err == "" {
		r.err = err.Error()
	}
}

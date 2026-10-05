// Package game is the contract between a player and a running game. Every
// player (an AI agent, a scripted baseline, a human at a terminal) plays
// through a Session: View the state, read the Decision, Choose an option.
// The same contract is exposed as JSON tools for language models (tools.go).
// Everything is deterministic: a scenario, seed and sequence of choices
// always produce the same views, decisions and results.
package game

import (
	"fmt"
	"math/rand/v2"
	"runtime/debug"

	"mcbench/engine"
	"mcbench/scenario"
)

// Status of a session.
const (
	AwaitingDecision = "awaiting_decision"
	Won              = "won"
	Lost             = "lost"
	RoundLimit       = "round_limit"    // max rounds played without a result
	DecisionLimit    = "decision_limit" // runaway game
	EngineError      = "engine_error"
)

// Decision is what the player must decide now.
type Decision struct {
	Kind    string   `json:"kind"` // mulligan, turn, defend, window, choice, discard
	Prompt  string   `json:"prompt"`
	Options []Option `json:"options"`
}

// Option is one legal choice. IDs are 1..n in the order shown. Key is a
// stable engine label (e.g. "play:Swinging Web Kick>Rhino|pay:Genius+Energy").
type Option struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Key  string `json:"key"`
}

// Choice records one decision made in the game.
type Choice struct {
	Round     int    `json:"round"`
	Kind      string `json:"kind"`
	Key       string `json:"key"`
	Reasoning string `json:"reasoning,omitempty"`
}

// Options configure a session.
type Options struct {
	MaxRounds    int
	MaxDecisions int
	// ShuffleOptions presents options in a seeded random order, so a
	// player's position bias cannot leak into results.
	ShuffleOptions bool
	// Sample varies the option order between games on the same seed.
	Sample uint64
}

// Session is one game in progress.
type Session struct {
	g      *engine.Game
	opts   Options
	stages int // villain stages at setup
	err    string

	rng      *rand.Rand
	order    []int // shown id - 1 -> engine option id, for the current decision
	orderFor int   // g.Decisions when order was made
	mark     int   // log length at the last choice

	Choices []Choice
	Invalid int // rejected choose_option calls
}

// New sets up a scenario game with a seed and runs it to the first decision.
func New(sc *scenario.Scenario, seed uint64, opts Options) (*Session, error) {
	g, err := sc.NewGame(seed)
	if err != nil {
		return nil, err
	}
	s := &Session{g: g, opts: opts, orderFor: -1, stages: 1 + len(g.S.VillainStages),
		rng: rand.New(rand.NewPCG(seed, opts.Sample+0x5eed))}
	g.OnPhase = func(p engine.Phase, round int) {
		if opts.MaxRounds > 0 && p == engine.PhasePlayer && round > opts.MaxRounds {
			g.Halt()
		}
	}
	s.guard(g.StartGame)
	return s, nil
}

// guard turns an engine panic into an engine_error status.
func (s *Session) guard(f func()) {
	defer func() {
		if r := recover(); r != nil {
			s.err = fmt.Sprintf("%v\n%s", r, debug.Stack())
		}
	}()
	f()
}

// Status reports whether the game awaits a decision or how it ended.
func (s *Session) Status() string {
	g := s.g
	switch {
	case s.err != "":
		return EngineError
	case g.Result().Over && g.Result().Won:
		return Won
	case g.Result().Over:
		return Lost
	case g.Halted():
		return RoundLimit
	case s.opts.MaxDecisions > 0 && g.Decisions >= s.opts.MaxDecisions:
		return DecisionLimit
	case g.Pending() == nil:
		s.err = "engine stopped without a decision or result"
		return EngineError
	}
	return AwaitingDecision
}

// Error is the engine error behind an engine_error status.
func (s *Session) Error() string { return s.err }

// View is the public game state.
func (s *Session) View() View { return buildView(s.g, s.stages) }

// Decision is the current decision, or nil when the game is over.
func (s *Session) Decision() *Decision {
	if s.Status() != AwaitingDecision {
		return nil
	}
	d := s.g.Pending()
	if s.orderFor != s.g.Decisions {
		s.orderFor = s.g.Decisions
		s.order = make([]int, len(d.Options))
		for i := range s.order {
			s.order[i] = i + 1
		}
		if s.opts.ShuffleOptions {
			s.rng.Shuffle(len(s.order), func(i, j int) { s.order[i], s.order[j] = s.order[j], s.order[i] })
		}
	}
	out := &Decision{Kind: d.Kind, Prompt: d.Prompt}
	for i, id := range s.order {
		o := d.Options[id-1]
		out.Options = append(out.Options, Option{ID: i + 1, Text: o.Text, Key: o.FullKey()})
	}
	return out
}

// Choose answers the current decision with a shown option id and runs the
// game to the next decision. It returns the events that happened.
func (s *Session) Choose(id int, reasoning string) ([]string, error) {
	d := s.Decision()
	if d == nil {
		return nil, fmt.Errorf("the game is over (%s)", s.Status())
	}
	if id < 1 || id > len(d.Options) {
		s.Invalid++
		return nil, fmt.Errorf("option %d does not exist; choose 1-%d", id, len(d.Options))
	}
	s.Choices = append(s.Choices, Choice{Round: s.g.S.Round, Kind: d.Kind, Key: d.Options[id-1].Key, Reasoning: reasoning})
	s.guard(func() { s.g.Choose(s.order[id-1]) })
	events := append([]string(nil), s.g.Log[s.mark:]...)
	s.mark = len(s.g.Log)
	return events, nil
}

// Log returns the last n lines of the game log (all when n <= 0).
func (s *Session) Log(n int) []string {
	log := s.g.Log
	if n > 0 && n < len(log) {
		log = log[len(log)-n:]
	}
	return append([]string(nil), log...)
}

// Stats are numeric facts about the game, used for scoring.
func (s *Session) Stats() map[string]float64 {
	m := s.g.Metrics()
	if s.opts.MaxRounds > 0 && m["round"] > float64(s.opts.MaxRounds) {
		m["round"] = float64(s.opts.MaxRounds)
	}
	return m
}

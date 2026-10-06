// Package session is the contract between a player and a running game. Every
// player plays through a Session: View the state, read the Decision, Choose
// an option. The model player exposes the same contract as JSON tools; see
// player/model. A scenario, seed and sequence of choices always produce the
// same game.
package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"runtime/debug"
	"time"

	"mcbench/constants"
	"mcbench/game/engine"
	"mcbench/game/scenario"
)

// Session is one game in progress.
type Session struct {
	g      *engine.Game
	opts   Options
	log    *slog.Logger
	stages int // villain stages at setup
	err    string

	rng      *rand.Rand
	order    []int // shown id - 1 -> engine option id, for the current decision
	orderFor int   // g.Decisions when order was made
	mark     int   // log length at the last choice

	Choices []Choice
	Invalid int // rejected choose_option calls
}

// sampleSalt mixes the sample number into the option-order RNG seed so
// different samples order options differently.
const sampleSalt = 0x5eed

// New sets up a scenario game with a seed and runs it to the first decision.
func New(sc *scenario.Scenario, seed uint64, opts Options) (*Session, error) {
	g, err := sc.NewGame(seed)
	if err != nil {
		return nil, err
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	s := &Session{g: g, opts: opts, log: logger, orderFor: -1, stages: 1 + len(g.S.VillainStages),
		rng: rand.New(rand.NewPCG(seed, opts.Sample+sampleSalt))}
	g.Logger = logger
	g.OnPhase = func(p engine.Phase, round int) {
		if opts.MaxRounds > 0 && p == engine.PhasePlayer && round > opts.MaxRounds {
			g.Halt()
		}
	}
	s.guard(g.StartGame)
	return s, nil
}

// Logger is the session's logger, already tagged with the game it belongs to.
func (s *Session) Logger() *slog.Logger { return s.log }

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
		return constants.EngineError
	case g.Result().Over && g.Result().Won:
		return constants.Won
	case g.Result().Over:
		return constants.Lost
	case g.Halted():
		return constants.RoundLimit
	case s.opts.MaxDecisions > 0 && g.Decisions >= s.opts.MaxDecisions:
		return constants.DecisionLimit
	case g.Pending() == nil:
		s.err = "engine stopped without a decision or result"
		return constants.EngineError
	}
	return constants.AwaitingDecision
}

// Error is the engine error behind an engine_error status.
func (s *Session) Error() string { return s.err }

// View is the public game state.
func (s *Session) View() View { return buildView(s.g, s.stages) }

// Decision is the current decision, or nil when the game is over.
func (s *Session) Decision() *Decision {
	if s.Status() != constants.AwaitingDecision {
		return nil
	}
	d := s.g.Pending()
	s.fixOrder(d)
	out := &Decision{Kind: d.Kind, Prompt: d.Prompt}
	for i, id := range s.order {
		o := d.Options[id-1]
		out.Options = append(out.Options, Option{ID: i + 1, Text: o.Text, Key: o.FullKey()})
	}
	return out
}

// fixOrder shuffles and numbers the options once per engine decision.
func (s *Session) fixOrder(d *engine.Decision) {
	if s.orderFor == s.g.Decisions {
		return
	}
	s.orderFor = s.g.Decisions
	s.order = make([]int, len(d.Options))
	for i := range s.order {
		s.order[i] = i + 1
	}
	if s.opts.ShuffleOptions {
		s.rng.Shuffle(len(s.order), func(i, j int) { s.order[i], s.order[j] = s.order[j], s.order[i] })
	}
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
		s.log.Warn("invalid choice", "option", id, "options", len(d.Options), "round", s.g.S.Round)
		return nil, fmt.Errorf("option %d does not exist; choose 1-%d", id, len(d.Options))
	}
	choice := Choice{Round: s.g.S.Round, Kind: d.Kind, Key: d.Options[id-1].Key, Reasoning: reasoning}
	s.Choices = append(s.Choices, choice)
	s.log.Info("choice", "n", len(s.Choices), "round", choice.Round, "kind", choice.Kind, "key", choice.Key)
	s.guard(func() { s.g.Choose(s.order[id-1]) })
	events := append([]string(nil), s.g.Log[s.mark:]...)
	s.mark = len(s.g.Log)
	s.Trace(constants.TraceChoice, &choiceTrace{
		Decision: len(s.Choices), Choice: choice, Text: d.Options[id-1].Text, Events: events, Status: s.Status(),
	})
	return events, nil
}

// TraceHeader is embedded in every trace entry; Trace fills it in.
type TraceHeader struct {
	Type string `json:"type"`
	Time string `json:"time"`
}

func (h *TraceHeader) header() *TraceHeader { return h }

// TraceEntry is a trace line: a struct embedding TraceHeader.
type TraceEntry interface{ header() *TraceHeader }

type choiceTrace struct {
	TraceHeader
	Decision int      `json:"decision"`
	Choice   Choice   `json:"choice"`
	Text     string   `json:"text"`
	Events   []string `json:"events"`
	Status   string   `json:"status"`
}

// Trace writes one JSON line of the given type to the session's trace, if it
// has one.
func (s *Session) Trace(typ string, entry TraceEntry) {
	if s.opts.Trace == nil {
		return
	}
	h := entry.header()
	h.Type, h.Time = typ, time.Now().Format(time.RFC3339)
	if b, err := json.Marshal(entry); err == nil {
		s.opts.Trace.Write(append(b, '\n'))
	}
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
	if s.opts.MaxRounds > 0 && m[constants.MetricRound] > float64(s.opts.MaxRounds) {
		m[constants.MetricRound] = float64(s.opts.MaxRounds)
	}
	return m
}

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
	"runtime/debug"
	"strings"
	"time"

	"mcbench/constants"
	"mcbench/game/engine"
	"mcbench/game/scenario"
)

// Session is one game in progress.
type Session struct {
	g      *engine.Game
	sc     *scenario.Scenario
	seed   uint64
	opts   Options
	log    *slog.Logger
	stages int // villain stages at setup
	err    string

	mark int // log length at the last choice

	Choices []Choice
	Invalid int // rejected choose_option calls
}

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
	s := &Session{g: g, sc: sc, seed: seed, opts: opts, log: logger, stages: 1 + len(g.S.VillainStages)}
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

// Decision is the current decision, or nil when the game is over. Options
// keep the engine's order, which groups them by kind of action (on a turn:
// card plays, hero powers, ally actions, card actions, form change, end turn).
func (s *Session) Decision() *Decision {
	if s.Status() != constants.AwaitingDecision {
		return nil
	}
	d := s.g.Pending()
	out := &Decision{Kind: d.Kind, Prompt: d.Prompt}
	for _, o := range d.Options {
		out.Options = append(out.Options, Option{ID: o.ID, Text: o.Text, Key: o.FullKey()})
	}
	return out
}

// Preview is the decision that follows option id, or nil when the game ends
// there. It changes nothing: the game is replayed from the start on a copy and
// the choice is made there, so it costs a replay per call.
func (s *Session) Preview(id int) (*Decision, error) {
	opts := s.opts
	opts.Trace, opts.Logger = nil, slog.New(slog.DiscardHandler)
	c, err := New(s.sc, s.seed, opts)
	if err != nil {
		return nil, err
	}
	for _, past := range s.Choices {
		if past.Auto && !s.opts.KeepForced {
			continue // the copy applies forced choices itself
		}
		if err := c.chooseKey(past.Key); err != nil {
			return nil, fmt.Errorf("replaying %q: %w", past.Key, err)
		}
	}
	if _, err := c.Choose(id, ""); err != nil {
		return nil, err
	}
	return c.Decision(), nil
}

// chooseKey makes the option with the given engine key.
func (s *Session) chooseKey(key string) error {
	d := s.Decision()
	if d == nil {
		return fmt.Errorf("the game is over")
	}
	for _, o := range d.Options {
		if o.Key == key {
			_, err := s.Choose(o.ID, "")
			return err
		}
	}
	return fmt.Errorf("%q is not offered", key)
}

// Choose answers the current decision with a shown option id and runs the
// game to the next decision that has more than one option: a decision with a
// single legal option is applied at once (unless Options.KeepForced), recorded
// as an Auto choice, and shown in the returned events as EventAutoResolved.
// It returns the events that happened.
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
	events := s.apply(d, id, reasoning, false)
	for !s.opts.KeepForced {
		d := s.Decision()
		if d == nil || len(d.Options) != 1 {
			break
		}
		forced := s.apply(d, 1, "", true)
		forced[0] = constants.EventAutoResolved + strings.TrimPrefix(forced[0], constants.EventChose)
		events = append(events, forced...)
	}
	return events, nil
}

// apply records and runs option id of d. The trace gets the engine's own log
// lines, so a replay that applies every recorded choice matches it.
func (s *Session) apply(d *Decision, id int, reasoning string, auto bool) []string {
	choice := s.choice(d, id, reasoning, auto)
	s.Choices = append(s.Choices, choice)
	s.log.Info("choice", "n", len(s.Choices), "round", choice.Round, "kind", choice.Kind, "key", choice.Key, "auto", auto)
	s.guard(func() { s.g.Choose(id) })
	events := append([]string(nil), s.g.Log[s.mark:]...)
	s.mark = len(s.g.Log)
	s.Trace(constants.TraceChoice, &choiceTrace{
		Decision: len(s.Choices), Choice: choice, Text: d.Options[id-1].Text, Events: events, Status: s.Status(),
	})
	return events
}

// choice records option id of d with the board as it is before it applies.
func (s *Session) choice(d *Decision, id int, reasoning string, auto bool) Choice {
	st := s.g.S
	form := constants.TypeAlterEgo
	if s.g.IsHero() {
		form = constants.TypeHero
	}
	offered := make([]string, len(d.Options))
	for i, o := range d.Options {
		offered[i] = o.Key
	}
	return Choice{
		Round: st.Round, Kind: d.Kind, Key: d.Options[id-1].Key, Reasoning: reasoning, Auto: auto,
		Form: form, HP: st.Hero.RemainingHP(), Threat: st.MainScheme.Threat, Offered: offered,
	}
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

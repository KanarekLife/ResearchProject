package engine

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

// Phase of the round.
type Phase string

const (
	PhasePlayer  Phase = "player"
	PhaseVillain Phase = "villain"
)

// State is the full game state, including hidden information. Agents never
// see State directly; they get an Observation.
type State struct {
	Round int
	Phase Phase

	Hero    *Card
	Hand    []*Card
	Deck    []*Card // index 0 is the top
	Discard []*Card
	Play    []*Card // allies, supports and unattached upgrades

	Villain       *Card
	VillainStages []*CardDef // stages after the current one, in order
	MainScheme    *Card
	SideSchemes   []*Card
	Minions       []*Card // engaged with the hero
	EncDeck       []*Card // index 0 is the top
	EncDiscard    []*Card
	AccelTokens   int
	// Encounter cards dealt face-down to the hero, revealed in the villain phase.
	Dealt []*Card

	// Nemesis holds your set-aside nemesis set (minion, side scheme, rest).
	Nemesis []*Card
	// Removed holds cards removed from the game.
	Removed []*Card

	// VillainTotalHP is the villain's HP over all stages at setup, for
	// measuring damage dealt.
	VillainTotalHP int

	// Per-turn / per-round bookkeeping.
	FormChanged  bool
	OncePerRound map[string]bool
	// CostReduction lowers the cost of the next card played this phase.
	CostReduction int
}

// Option is one legal choice in a decision.
type Option struct {
	ID int
	// Key is a stable, engine-generated label for grading, e.g.
	// "play:Swinging Web Kick>Rhino" or "attack:Spider-Man>Rhino".
	Key string
	// Pay describes the resources spent, e.g. "Energy+Genius". Empty if free.
	Pay string
	// Text is the human-readable description shown to the agent.
	Text string
	do   func()
}

// FullKey is Key plus payment, for exact grading.
func (o Option) FullKey() string {
	if o.Pay == "" {
		return o.Key
	}
	return o.Key + "|pay:" + o.Pay
}

// Decision is a prompt the player has to answer.
type Decision struct {
	Kind    string // "turn", "defend", "window", "choice", "discard"
	Prompt  string
	Options []Option
}

// Result of a finished game.
type Result struct {
	Over   bool
	Won    bool
	Reason string
}

// Game drives the rules. It runs scheduled steps until it needs a decision
// from the player or the game ends.
type Game struct {
	S   *State
	rng *rand.Rand

	stack   []func()
	pending *Decision
	result  Result

	// LastPaid holds the resources spent on the most recent payment, for
	// cards that check "if you paid with a [mental] resource".
	LastPaid []Resource

	Log       []string
	Decisions int // decisions answered so far

	// OnPhase, if set, is called whenever a player turn or a villain phase
	// begins. Harnesses use it with Halt to stop at a horizon.
	OnPhase func(phase Phase, round int)
	halted  bool
}

// Halt stops Run before the next step. The game cannot be resumed.
func (g *Game) Halt() { g.halted = true }

// Halted reports whether Halt was called.
func (g *Game) Halted() bool { return g.halted }

func (g *Game) phaseBegins(p Phase) {
	if g.OnPhase != nil {
		g.OnPhase(p, g.S.Round)
	}
}

// New creates a game around a prepared state. The seed drives every random
// operation (shuffles); identical seeds and choices give identical games.
func New(s *State, seed uint64) *Game {
	g := &Game{S: s, rng: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))}
	if s.OncePerRound == nil {
		s.OncePerRound = map[string]bool{}
	}
	return g
}

// Do schedules steps so that steps[0] runs next.
func (g *Game) Do(steps ...func()) {
	for i := len(steps) - 1; i >= 0; i-- {
		g.stack = append(g.stack, steps[i])
	}
}

// Ask pauses the game on a decision. Options are numbered here.
func (g *Game) Ask(d *Decision) {
	if len(d.Options) == 0 {
		panic("engine: decision without options: " + d.Prompt)
	}
	for i := range d.Options {
		d.Options[i].ID = i + 1
	}
	g.pending = d
}

// Run advances until a decision is pending or the game is over.
func (g *Game) Run() {
	for !g.result.Over && !g.halted && g.pending == nil && len(g.stack) > 0 {
		step := g.stack[len(g.stack)-1]
		g.stack = g.stack[:len(g.stack)-1]
		step()
	}
}

// Pending returns the open decision, or nil.
func (g *Game) Pending() *Decision { return g.pending }

// Result returns the game result (Over is false while the game runs).
func (g *Game) Result() Result { return g.result }

// Choose answers the pending decision with an option ID and advances.
func (g *Game) Choose(id int) error {
	d := g.pending
	if d == nil || g.halted {
		return errors.New("no decision is pending")
	}
	if id < 1 || id > len(d.Options) {
		return fmt.Errorf("option %d does not exist (valid: 1-%d)", id, len(d.Options))
	}
	opt := d.Options[id-1]
	g.pending = nil
	g.Decisions++
	g.Logf("Player chose: %s", opt.Text)
	if opt.do != nil {
		opt.do()
	}
	g.Run()
	return nil
}

// Logf appends a line to the public game log.
func (g *Game) Logf(format string, args ...any) {
	g.Log = append(g.Log, fmt.Sprintf(format, args...))
}

func (g *Game) lose(reason string) {
	if g.result.Over {
		return
	}
	g.result = Result{Over: true, Won: false, Reason: reason}
	g.Logf("GAME OVER - defeat: %s", reason)
	g.stack = nil
	g.pending = nil
}

func (g *Game) win(reason string) {
	if g.result.Over {
		return
	}
	g.result = Result{Over: true, Won: true, Reason: reason}
	g.Logf("GAME OVER - victory: %s", reason)
	g.stack = nil
	g.pending = nil
}

func (g *Game) shuffle(cards []*Card) {
	g.rng.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
}

// inPlay visits every card in play (identity, villain, schemes, minions,
// player cards and attachments).
func (g *Game) inPlay(f func(*Card)) {
	s := g.S
	var visit func(c *Card)
	visit = func(c *Card) {
		if c == nil {
			return
		}
		f(c)
		for _, a := range c.Attached {
			visit(a)
		}
	}
	visit(s.Hero)
	visit(s.Villain)
	visit(s.MainScheme)
	for _, zone := range [][]*Card{s.Play, s.SideSchemes, s.Minions} {
		for _, c := range zone {
			visit(c)
		}
	}
}

// IsHero reports whether the identity is in hero form.
func (g *Game) IsHero() bool { return g.S.Hero.Face().Type == TypeHero }

func remove(zone []*Card, c *Card) []*Card {
	for i, x := range zone {
		if x == c {
			return append(zone[:i:i], zone[i+1:]...)
		}
	}
	return zone
}

func contains(zone []*Card, c *Card) bool {
	for _, x := range zone {
		if x == c {
			return true
		}
	}
	return false
}

// Shuffle shuffles cards in place with the game's seeded RNG.
func (g *Game) Shuffle(cards []*Card) { g.shuffle(cards) }

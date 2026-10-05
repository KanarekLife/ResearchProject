package engine

import "mcbench/constants"

// Phase of the round.
type Phase string

const (
	PhasePlayer  Phase = "player"
	PhaseVillain Phase = "villain"
)

// State is the full game state, including hidden information. Players never
// see State directly; they get a view from the session package.
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
	// Text is the human-readable description shown to the player.
	Text string
	do   func()
}

// FullKey is Key plus payment, for exact grading.
func (o Option) FullKey() string {
	if o.Pay == "" {
		return o.Key
	}
	return o.Key + constants.SepPayment + o.Pay
}

// Decision is a prompt the player has to answer.
type Decision struct {
	Kind    string // turn, defend, window, choice, discard
	Prompt  string
	Options []Option
}

// Result of a finished game.
type Result struct {
	Over   bool
	Won    bool
	Reason string
}

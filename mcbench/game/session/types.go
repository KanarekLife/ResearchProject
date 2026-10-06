package session

import (
	"io"
	"log/slog"
)

// Decision is what the player must decide now.
type Decision struct {
	Kind    string   `json:"kind"` // mulligan, turn, defend, window, choice, discard
	Prompt  string   `json:"prompt"`
	Options []Option `json:"options"`
}

// Option is one legal choice. IDs are 1..n in the order shown. Key is a
// stable engine label, e.g. "play:Swinging Web Kick>Rhino|pay:Genius+Energy".
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
	// Trace, if set, receives a JSON line for every choice, plus anything
	// players add with Session.Trace (e.g. model messages).
	Trace io.Writer
	// Logger receives game actions and choices at info level. Nil uses
	// slog.Default.
	Logger *slog.Logger
}

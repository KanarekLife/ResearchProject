package session

import (
	"io"
	"log/slog"
)

// Decision is what the player must decide now.
type Decision struct {
	Kind    string   `json:"kind"` // mulligan, turn, defend, window, choice, discard, payment
	Prompt  string   `json:"prompt"`
	Options []Option `json:"options"`
}

// Option is one legal choice. IDs are 1..n in the order shown. Key is a
// stable engine label, e.g. "play:Swinging Web Kick>Rhino" or "pay:Genius+Energy".
type Option struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Key  string `json:"key"`
}

// Choice records one decision made in the game, with the board facts and
// options the player had when making it (for behaviour analysis).
type Choice struct {
	Round     int      `json:"round"`
	Kind      string   `json:"kind"`
	Key       string   `json:"key"`
	Reasoning string   `json:"reasoning,omitempty"`
	Auto      bool     `json:"auto,omitempty"` // forced: the only legal option, applied by the session
	Form      string   `json:"form,omitempty"` // hero or alter-ego
	HP        int      `json:"hp"`             // the hero's remaining hit points
	Threat    int      `json:"threat"`         // on the main scheme
	Offered   []string `json:"offered,omitempty"`
}

// Options configure a session.
type Options struct {
	MaxRounds    int
	MaxDecisions int
	// KeepForced stops the session from applying a decision's only legal option
	// itself. A replay sets it so that it applies every recorded choice.
	KeepForced bool
	// Trace, if set, receives a JSON line for every choice, plus anything
	// players add with Session.Trace (e.g. model messages).
	Trace io.Writer
	// Logger receives game actions and choices at info level. Nil uses
	// slog.Default.
	Logger *slog.Logger
}

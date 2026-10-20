// Package player is the player side of the benchmark. A Player answers the
// decisions of one game through the session contract alone. Implementations
// live in subpackages: heuristic (scripted), human, and model (an LLM driven
// through the inference client).
package player

import (
	"context"
	"errors"

	"mcbench/game/session"
)

// ErrInference marks a game a player could not finish because the model could
// not be reached (network, HTTP status, a bad response or a cancelled run).
// It says nothing about how the model plays, so such a game is not graded.
var ErrInference = errors.New("inference failed")

// Player plays one game to its end.
type Player interface {
	// Name identifies the player configuration in results.
	Name() string
	// Play makes decisions until the session is no longer awaiting one. It
	// must be safe to call concurrently for different sessions.
	Play(ctx context.Context, s *session.Session) (Usage, error)
}

// Usage counts model tokens and requests.
type Usage struct {
	Requests     int   `json:"requests,omitempty"`
	InputTokens  int64 `json:"input_tokens,omitempty"`
	OutputTokens int64 `json:"output_tokens,omitempty"`
	CachedTokens int64 `json:"cached_tokens,omitempty"`
}

type gameKey struct{}

// Game names the run and game being played. A model player forwards it to its
// inference client, which fills request-header placeholders from it.
type Game struct{ Run, Name string }

// WithRun and WithGame set the parts of the Game carried by ctx.
func WithRun(ctx context.Context, run string) context.Context {
	g := GameFrom(ctx)
	g.Run = run
	return context.WithValue(ctx, gameKey{}, g)
}

func WithGame(ctx context.Context, name string) context.Context {
	g := GameFrom(ctx)
	g.Name = name
	return context.WithValue(ctx, gameKey{}, g)
}

func GameFrom(ctx context.Context) Game {
	g, _ := ctx.Value(gameKey{}).(Game)
	return g
}

type resumeKey struct{}

// WithResume hands a player the trace of a game it is to continue.
func WithResume(ctx context.Context, trace []byte) context.Context {
	return context.WithValue(ctx, resumeKey{}, trace)
}

// ResumeFrom is the trace of the game to continue, or nil for a new game.
func ResumeFrom(ctx context.Context) []byte {
	t, _ := ctx.Value(resumeKey{}).([]byte)
	return t
}

// Package player is the player side of the benchmark. A Player answers the
// decisions of one game through the session contract alone. Implementations
// live in subpackages: heuristic (scripted), human, and model (an LLM driven
// through the inference client).
package player

import (
	"context"

	"mcbench/game/session"
)

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

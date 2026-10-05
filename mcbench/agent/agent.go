// Package agent contains players. Each plays a whole game through the
// game.Session contract and nothing else: language models (llm.go), scripted
// baselines (baselines.go) and a human at a terminal (human.go).
package agent

import (
	"context"

	"mcbench/game"
)

// Agent plays one game to its end.
type Agent interface {
	// Name identifies the player configuration in results.
	Name() string
	// Play makes decisions until the session is no longer awaiting one. It
	// must be safe to call concurrently for different sessions.
	Play(ctx context.Context, s *game.Session) (Usage, error)
}

// Usage counts model tokens and requests.
type Usage struct {
	Requests     int   `json:"requests,omitempty"`
	InputTokens  int64 `json:"input_tokens,omitempty"`
	OutputTokens int64 `json:"output_tokens,omitempty"`
}

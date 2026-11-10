package inference

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Placeholders a header value may hold. They are filled per request from the
// game being played, so one run can send a different value for every game.
const (
	PlaceholderRun  = "{run}"  // the run's directory name, unique per run
	PlaceholderGame = "{game}" // scenario, seed and sample, e.g. spider-man-vs-rhino_seed2_s0
)

var placeholder = regexp.MustCompile(`\{[^{}]*\}`)

// ValidateHeaderValue rejects a placeholder that is not one of the above.
func ValidateHeaderValue(value string) error {
	for _, p := range placeholder.FindAllString(value, -1) {
		if p != PlaceholderRun && p != PlaceholderGame {
			return fmt.Errorf("unknown placeholder %s: use %s or %s", p, PlaceholderRun, PlaceholderGame)
		}
	}
	return nil
}

type varsKey struct{}

type gameVars struct{ run, game string }

// WithGame returns a context that fills the placeholders with this run and game.
func WithGame(ctx context.Context, run, game string) context.Context {
	return context.WithValue(ctx, varsKey{}, gameVars{run, game})
}

// GameFrom is the run and game set by WithGame, empty outside a game.
func GameFrom(ctx context.Context) (run, game string) {
	v, _ := ctx.Value(varsKey{}).(gameVars)
	return v.run, v.game
}

// ExpandHeaders returns headers with the placeholders filled from ctx; outside
// a game they are left as written.
func ExpandHeaders(ctx context.Context, headers map[string]string) map[string]string {
	v, ok := ctx.Value(varsKey{}).(gameVars)
	if !ok {
		return headers
	}
	r := strings.NewReplacer(PlaceholderRun, v.run, PlaceholderGame, v.game)
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		out[k] = r.Replace(v)
	}
	return out
}

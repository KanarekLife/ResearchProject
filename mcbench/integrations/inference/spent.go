package inference

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrPlanSpent is a request refused because the provider's plan is spent (a
// monthly limit, no credits left). Waiting or retrying cannot help, so the run
// stops instead of failing its remaining games one by one.
var ErrPlanSpent = errors.New("provider plan spent")

type spentKey struct{}

type spentPlan struct {
	texts   []string
	onSpent func(error)
}

// WithSpentPlan makes a request whose error contains one of texts fail at once
// with ErrPlanSpent, after calling onSpent with that error.
func WithSpentPlan(ctx context.Context, texts []string, onSpent func(error)) context.Context {
	if len(texts) == 0 {
		return ctx
	}
	return context.WithValue(ctx, spentKey{}, spentPlan{texts, onSpent})
}

// planSpent is err as ErrPlanSpent when ctx names it a spent plan, else nil.
func planSpent(ctx context.Context, err error) error {
	sp, ok := ctx.Value(spentKey{}).(spentPlan)
	if !ok {
		return nil
	}
	msg := err.Error()
	for _, t := range sp.texts {
		if strings.Contains(msg, t) {
			err = fmt.Errorf("%w: %w", ErrPlanSpent, err)
			sp.onSpent(err)
			return err
		}
	}
	return nil
}

// Package systemone plays a game with a System One classifier: for every
// decision it sends the board and the legal options, and takes the option the
// model rates most likely. There is no conversation, so each decision sees only
// the board and the last few log lines.
package systemone

import (
	"context"
	"fmt"
	"strconv"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/integrations/inference"
	"mcbench/integrations/systemone"
	"mcbench/player"
	"mcbench/player/instruction"
	"mcbench/player/model"
)

// instructions is the question asked at every decision.
const instructions = "You are playing a solo game of Marvel Champions: The Card Game. " +
	"Win by defeating the villain before the main scheme's threat reaches its threshold " +
	"and before your hero is defeated. Pick the best option for the decision."

// SystemOne is the classifier player.
type SystemOne struct {
	Client *systemone.Client
	Label  string // identifies the model in results
	// RecentEvents is how many log lines each decision carries, since the
	// model remembers nothing between decisions (0: the whole log).
	RecentEvents int
	// Instructions are appended to the question; none by default.
	Instructions instruction.Instructions
}

func (p *SystemOne) Name() string { return "systemone:" + p.Label }

// request is the state sent with every decision.
type request struct {
	Board    any      `json:"board"`
	Recent   []string `json:"recent_events,omitempty"`
	Decision string   `json:"decision"`
	Kind     string   `json:"kind"`
}

// question is the instruction sent with every decision.
func (p *SystemOne) question() string {
	if p.Instructions.Text == "" {
		return instructions
	}
	return instructions + "\n\n" + p.Instructions.Text
}

// Play answers every decision with the option rated most likely.
func (p *SystemOne) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	var u player.Usage
	g := player.GameFrom(ctx)
	ctx = inference.WithGame(ctx, g.Run, g.Name)
	for s.Status() == constants.AwaitingDecision {
		d := s.Decision()
		opts := make([]systemone.Option, len(d.Options))
		for i, o := range d.Options {
			opts[i] = systemone.Option{Name: strconv.Itoa(o.ID), Description: o.Text}
		}
		state := request{Board: model.BoardState(s), Recent: s.Log(p.RecentEvents), Decision: d.Prompt, Kind: d.Kind}
		ans, err := p.Client.Choose(ctx, state, p.question(), opts)
		u.Requests++
		u.InputTokens += ans.InputTokens
		if err != nil {
			return u, fmt.Errorf("%w: %w", player.ErrInference, err)
		}
		id, err := strconv.Atoi(ans.Choice)
		if err != nil {
			return u, fmt.Errorf("%w: system one answered %q, not an option id", player.ErrInference, ans.Choice)
		}
		if _, err := s.Choose(id, fmt.Sprintf("p=%.2f", ans.Probabilities[ans.Choice])); err != nil {
			return u, fmt.Errorf("%w: %w", player.ErrInference, err)
		}
	}
	return u, nil
}

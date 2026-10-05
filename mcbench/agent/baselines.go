package agent

import (
	"context"
	"math/rand/v2"
	"strings"

	"mcbench/game"
)

// Random picks uniformly at random: the chance baseline.
type Random struct{ Seed uint64 }

func (r Random) Name() string { return "random" }

func (r Random) Play(ctx context.Context, s *game.Session) (Usage, error) {
	rng := rand.New(rand.NewPCG(r.Seed, 7))
	return Usage{}, playEach(ctx, s, func(d *game.Decision) int { return 1 + rng.IntN(len(d.Options)) })
}

// First always picks option 1: a position-bias baseline.
type First struct{}

func (First) Name() string { return "first" }

func (First) Play(ctx context.Context, s *game.Session) (Usage, error) {
	return Usage{}, playEach(ctx, s, func(*game.Decision) int { return 1 })
}

// Heuristic is a scripted baseline for "competent simple play": keep the
// opening hand, flip to hero when healthy, thwart when the main scheme is
// close to its threshold, otherwise develop (allies, upgrades, supports) and
// attack. It always defends with the hero and plays interrupts.
type Heuristic struct{}

func (Heuristic) Name() string { return "heuristic" }

func (Heuristic) Play(ctx context.Context, s *game.Session) (Usage, error) {
	return Usage{}, playEach(ctx, s, func(d *game.Decision) int { return heuristicPick(s.View(), d) })
}

// playEach answers every decision with pick until the game ends.
func playEach(ctx context.Context, s *game.Session, pick func(*game.Decision) int) error {
	for d := s.Decision(); d != nil; d = s.Decision() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := s.Choose(pick(d), ""); err != nil {
			return err
		}
	}
	return nil
}

func heuristicPick(v game.View, d *game.Decision) int {
	find := func(prefixes ...string) int {
		for _, p := range prefixes {
			for _, o := range d.Options {
				if strings.HasPrefix(o.Key, p) {
					return o.ID
				}
			}
		}
		return 0
	}
	or := func(id int) int {
		if id == 0 {
			return d.Options[0].ID
		}
		return id
	}
	hero := v.Hero
	switch d.Kind {
	case "mulligan":
		return or(find("mulligan:keep"))
	case "discard":
		return or(find("discard:done"))
	case "defend":
		return or(find("defend:"+hero.Name, "defend:none"))
	case "window":
		if strings.Contains(d.Prompt, "threat is about to be placed") && hero.HP <= 6 {
			return or(find("pass"))
		}
		for _, o := range d.Options {
			if o.Key != "pass" {
				return o.ID
			}
		}
	case "turn":
		danger := v.MainScheme.Threat+3 >= v.MainScheme.Threshold
		if hero.Form == "alter-ego" {
			if hero.HP > 5 || danger {
				if id := find("change_form:"); id != 0 {
					return id
				}
			}
			return or(find("recover:", "use:Aunt May", "end_turn"))
		}
		if hero.HP <= 3 && !danger {
			if id := find("change_form:"); id != 0 {
				return id
			}
		}
		thwart := []string{"play:For Justice!>" + v.MainScheme.Name, "thwart:", "use:Surveillance Team"}
		develop := []string{"play:Web-Shooter", "play:Heroic Intuition", "play:Daredevil", "play:Jessica Jones", "play:Black Cat",
			"play:Mockingbird", "play:Nick Fury", "play:Aunt May", "play:Surveillance Team", "play:Interrogation Room",
			"play:Helicarrier", "use:Helicarrier", "use:Avengers Mansion", "play:Avengers Mansion"}
		attack := []string{"play:Swinging Web Kick>", "play:Haymaker>", "attack:"}
		order := append(append(develop, attack...), thwart...)
		if danger {
			order = append(thwart, attack...)
		}
		return or(find(append(order, "end_turn")...))
	}
	return d.Options[0].ID
}

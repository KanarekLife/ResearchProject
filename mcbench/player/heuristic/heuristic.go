// Package heuristic holds the scripted players used as reference points:
// Heuristic ("competent simple play"), Random (chance) and First (position
// bias). All are deterministic and need no model.
package heuristic

import (
	"context"
	"math/rand/v2"
	"strings"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/player"
)

// Card names the Heuristic builds its action order from.
const (
	cardForJustice        = "For Justice!"
	cardSurveillanceTeam  = "Surveillance Team"
	cardWebShooter        = "Web-Shooter"
	cardHeroicIntuition   = "Heroic Intuition"
	cardDaredevil         = "Daredevil"
	cardJessicaJones      = "Jessica Jones"
	cardBlackCat          = "Black Cat"
	cardMockingbird       = "Mockingbird"
	cardNickFury          = "Nick Fury"
	cardAuntMay           = "Aunt May"
	cardInterrogationRoom = "Interrogation Room"
	cardHelicarrier       = "Helicarrier"
	cardAvengersMansion   = "Avengers Mansion"
	cardSwingingWebKick   = "Swinging Web Kick"
	cardHaymaker          = "Haymaker"
)

// Heuristic's thresholds.
const (
	// windowLowHP: at or below this HP, skip interrupt events in a
	// threat-placement window, since they would turn threat into damage.
	windowLowHP = 6
	// alterEgoHealthyHP: above this in alter-ego form, flip to hero.
	alterEgoHealthyHP = 5
	// dangerMargin: threat this close to the threshold is danger.
	dangerMargin = 3
	// retreatHP: below this in hero form (and not in danger), flip back.
	retreatHP = 3
)

// threatPlacementPrompt matches the engine's threat-placement window prompt.
const threatPlacementPrompt = "threat is about to be placed"

// randomSalt mixes the player seed into the Random player's RNG.
const randomSalt = 7

// Random picks uniformly at random: the chance baseline.
type Random struct{ Seed uint64 }

func (Random) Name() string { return "random" }

func (r Random) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	rng := rand.New(rand.NewPCG(r.Seed, randomSalt))
	return player.Usage{}, playEach(ctx, s, func(d *session.Decision) int { return 1 + rng.IntN(len(d.Options)) })
}

// First always picks option 1: a position-bias baseline.
type First struct{}

func (First) Name() string { return "first" }

func (First) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	return player.Usage{}, playEach(ctx, s, func(*session.Decision) int { return 1 })
}

// Heuristic is a scripted baseline for "competent simple play": keep the
// opening hand, flip to hero when healthy, thwart when the main scheme is
// close to its threshold, otherwise develop (allies, upgrades, supports) and
// attack. It always defends with the hero and plays interrupts.
type Heuristic struct{}

func (Heuristic) Name() string { return "heuristic" }

func (Heuristic) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
	return player.Usage{}, playEach(ctx, s, func(d *session.Decision) int { return pick(s.View(), d) })
}

// playEach answers every decision with pick until the game ends.
func playEach(ctx context.Context, s *session.Session, pick func(*session.Decision) int) error {
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

func pick(v session.View, d *session.Decision) int {
	switch d.Kind {
	case constants.KindMulligan:
		return firstOf(d, optionWithKey(d, constants.KeyMulliganKeep))
	case constants.KindDiscard:
		return firstOf(d, optionWithKey(d, constants.KeyDiscardDone))
	case constants.KindDefend:
		return firstOf(d, optionWithKey(d, constants.PrefixDefend+v.Hero.Name, constants.PrefixDefend+"none"))
	case constants.KindWindow:
		return pickWindow(v, d)
	case constants.KindTurn:
		return pickTurn(v, d)
	}
	return d.Options[0].ID
}

// pickWindow plays the first event unless it is the low-HP threat window.
func pickWindow(v session.View, d *session.Decision) int {
	if strings.Contains(d.Prompt, threatPlacementPrompt) && v.Hero.HP <= windowLowHP {
		return firstOf(d, optionWithKey(d, constants.KeyPass))
	}
	for _, o := range d.Options {
		if o.Key != constants.KeyPass {
			return o.ID
		}
	}
	return d.Options[0].ID
}

// pickTurn chooses between flipping form, developing, attacking and thwarting.
func pickTurn(v session.View, d *session.Decision) int {
	danger := v.MainScheme.Threat+dangerMargin >= v.MainScheme.Threshold
	if v.Hero.Form == constants.TypeAlterEgo {
		return pickAlterEgoTurn(v, d, danger)
	}
	return pickHeroTurn(v, d, danger)
}

func pickAlterEgoTurn(v session.View, d *session.Decision, danger bool) int {
	if v.Hero.HP > alterEgoHealthyHP || danger {
		if id := optionWithKey(d, constants.PrefixChangeForm); id != 0 {
			return id
		}
	}
	return firstOf(d, optionWithKey(d,
		constants.PrefixRecover,
		constants.PrefixUse+cardAuntMay,
		constants.KeyEndTurn,
	))
}

func pickHeroTurn(v session.View, d *session.Decision, danger bool) int {
	if v.Hero.HP <= retreatHP && !danger {
		if id := optionWithKey(d, constants.PrefixChangeForm); id != 0 {
			return id
		}
	}
	thwart := []string{
		constants.PrefixPlay + cardForJustice + constants.SepTarget + v.MainScheme.Name,
		constants.PrefixThwart,
		constants.PrefixUse + cardSurveillanceTeam,
	}
	develop := []string{
		constants.PrefixPlay + cardWebShooter, constants.PrefixPlay + cardHeroicIntuition,
		constants.PrefixPlay + cardDaredevil, constants.PrefixPlay + cardJessicaJones, constants.PrefixPlay + cardBlackCat,
		constants.PrefixPlay + cardMockingbird, constants.PrefixPlay + cardNickFury, constants.PrefixPlay + cardAuntMay,
		constants.PrefixPlay + cardSurveillanceTeam, constants.PrefixPlay + cardInterrogationRoom,
		constants.PrefixPlay + cardHelicarrier, constants.PrefixUse + cardHelicarrier,
		constants.PrefixUse + cardAvengersMansion, constants.PrefixPlay + cardAvengersMansion,
	}
	attack := []string{
		constants.PrefixPlay + cardSwingingWebKick + constants.SepTarget,
		constants.PrefixPlay + cardHaymaker + constants.SepTarget,
		constants.PrefixAttack,
	}
	order := append(append(develop, attack...), thwart...)
	if danger {
		order = append(thwart, attack...)
	}
	return firstOf(d, optionWithKey(d, append(order, constants.KeyEndTurn)...))
}

// optionWithKey returns the id of the first option whose key starts with one
// of the given prefixes, or 0 when none match.
func optionWithKey(d *session.Decision, prefixes ...string) int {
	for _, p := range prefixes {
		for _, o := range d.Options {
			if strings.HasPrefix(o.Key, p) {
				return o.ID
			}
		}
	}
	return 0
}

// firstOf returns id, or the first option's id when id is 0.
func firstOf(d *session.Decision, id int) int {
	if id == 0 {
		return d.Options[0].ID
	}
	return id
}

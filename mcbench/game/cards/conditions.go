package cards

import (
	"slices"
	"strings"

	"mcbench/constants"
	e "mcbench/game/engine"
)

func when(g *e.Game, c *e.Card, ev *e.Event, preds []string) bool {
	for _, p := range preds {
		if !pred(g, c, ev, p) {
			return false
		}
	}
	return true
}

// predicates are the named conditions card YAML can use in `when`. The
// paid:<resource> and not_paid:<resource> families are handled separately by
// predicate.
var predicates = map[string]func(g *e.Game, c *e.Card, ev *e.Event) bool{
	constants.PredHero:             func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.IsHero() },
	constants.PredAlterEgo:         func(g *e.Game, c *e.Card, ev *e.Event) bool { return !g.IsHero() },
	constants.PredHeroDamaged:      func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.S.Hero.Damage > 0 },
	constants.PredHeroConfused:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.S.Hero.Confused },
	constants.PredVillainTough:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.S.Villain.Tough },
	constants.PredSchemes:          func(g *e.Game, c *e.Card, ev *e.Event) bool { return len(g.ThwartableSchemes()) > 0 },
	constants.PredMinions:          func(g *e.Game, c *e.Card, ev *e.Event) bool { return len(g.S.Minions) > 0 },
	constants.PredNotExhausted:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return c != nil && !c.Exhausted },
	constants.PredExhausted:        func(g *e.Game, c *e.Card, ev *e.Event) bool { return c != nil && c.Exhausted },
	constants.PredCountersPositive: func(g *e.Game, c *e.Card, ev *e.Event) bool { return c != nil && c.Counters > 0 },
	constants.PredHeroExhausted:    func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.S.Hero.Exhausted },
	constants.PredHeroReady:        func(g *e.Game, c *e.Card, ev *e.Event) bool { return !g.S.Hero.Exhausted },
	constants.PredSourceSelf:       func(g *e.Game, c *e.Card, ev *e.Event) bool { return ev != nil && ev.Source == c },
	constants.PredHostIsSource:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return ev != nil && g.HostOf(c) == ev.Source },
	constants.PredHostIsTarget:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return ev != nil && g.HostOf(c) == ev.Target },
	constants.PredBombScare: func(g *e.Game, c *e.Card, ev *e.Event) bool {
		return findCode(g.S.SideSchemes, constants.CodeBombScare) != nil
	},
	constants.PredNotBombScare: func(g *e.Game, c *e.Card, ev *e.Event) bool {
		return findCode(g.S.SideSchemes, constants.CodeBombScare) == nil
	},
	constants.PredVulture: func(g *e.Game, c *e.Card, ev *e.Event) bool {
		return findCode(g.S.Minions, constants.CodeVulture) != nil
	},
	constants.PredUpgradesSupports: func(g *e.Game, c *e.Card, ev *e.Event) bool { return len(controlledUpgradesSupports(g)) > 0 },
	constants.PredAmountPositive:   func(g *e.Game, c *e.Card, ev *e.Event) bool { return ev != nil && ev.Amount > 0 },
}

func predicate(p string) func(g *e.Game, c *e.Card, ev *e.Event) bool {
	if f, ok := predicates[p]; ok {
		return f
	}
	if r, ok := strings.CutPrefix(p, constants.PredPaidPrefix); ok {
		res := resource(r)
		return func(g *e.Game, _ *e.Card, _ *e.Event) bool { return paidWith(g, res) }
	}
	if r, ok := strings.CutPrefix(p, constants.PredNotPaidPrefix); ok {
		res := resource(r)
		return func(g *e.Game, _ *e.Card, _ *e.Event) bool { return !paidWith(g, res) }
	}
	return nil
}

// paidWith reports whether the last payment included res. A wild resource
// spent on a cost may be declared as any type.
func paidWith(g *e.Game, res e.Resource) bool {
	return slices.Contains(g.LastPaid, res) || slices.Contains(g.LastPaid, e.Wild)
}

func pred(g *e.Game, c *e.Card, ev *e.Event, p string) bool {
	f := predicate(p)
	if f == nil {
		panic("cards: unknown predicate " + p)
	}
	return f(g, c, ev)
}

func effectTarget(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) *e.Card {
	switch eff.Target {
	case constants.SelEventSource:
		if ev != nil {
			return ev.Source
		}
	case constants.SelEventTarget:
		if ev != nil {
			return ev.Target
		}
	case constants.SelChosen:
		return target
	case constants.SelHero:
		return g.S.Hero
	case constants.SelVillain:
		return g.S.Villain
	case constants.SelSelf:
		return c
	}
	return c
}

func selectTargets(g *e.Game, sel string) []*e.Card {
	switch sel {
	case constants.SelEnemy:
		return g.Enemies()
	case constants.SelAllEnemy:
		return g.AllEnemies()
	case constants.SelMinion:
		return append([]*e.Card(nil), g.S.Minions...)
	case constants.SelScheme:
		return g.ThwartableSchemes()
	case constants.SelUpgradeSupport:
		return controlledUpgradesSupports(g)
	case constants.SelEnemyWithoutWebbed:
		return enemiesWithoutWebbedUp(g)
	case constants.SelDamagedCharacter:
		return damagedCharacters(g)
	}
	return nil
}

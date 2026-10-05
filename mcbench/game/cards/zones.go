package cards

import (
	"mcbench/constants"
	e "mcbench/game/engine"
)

// controlledUpgradesSupports lists upgrades and supports you control,
// including your upgrades attached to enemies.
func controlledUpgradesSupports(g *e.Game) []*e.Card {
	var out []*e.Card
	isUS := func(c *e.Card) bool { return c.Def.Type == e.TypeUpgrade || c.Def.Type == e.TypeSupport }
	for _, c := range g.S.Play {
		if isUS(c) {
			out = append(out, c)
		}
	}
	for _, host := range append(g.AllEnemies(), g.S.Hero) {
		for _, a := range host.Attached {
			if isUS(a) {
				out = append(out, a)
			}
		}
	}
	return out
}

func enemiesWithoutWebbedUp(g *e.Game) []*e.Card {
	var out []*e.Card
	for _, en := range g.AllEnemies() {
		if findCode(en.Attached, constants.CodeWebbedUp) == nil {
			out = append(out, en)
		}
	}
	return out
}

func findCode(zone []*e.Card, code string) *e.Card {
	for _, c := range zone {
		if c.Def.Code == code {
			return c
		}
	}
	return nil
}

func removeCard(zone []*e.Card, c *e.Card) []*e.Card {
	out := make([]*e.Card, 0, len(zone))
	for _, x := range zone {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

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
	for _, host := range append(append(g.AllEnemies(), g.S.Hero), allies(g)...) {
		for _, a := range host.Attached {
			if isUS(a) {
				out = append(out, a)
			}
		}
	}
	return out
}

// controlledUpgrades lists the upgrades you control.
func controlledUpgrades(g *e.Game) []*e.Card {
	var out []*e.Card
	for _, c := range controlledUpgradesSupports(g) {
		if c.Def.Type == e.TypeUpgrade {
			out = append(out, c)
		}
	}
	return out
}

// allies lists the allies you control.
func allies(g *e.Game) []*e.Card {
	var out []*e.Card
	for _, c := range g.S.Play {
		if c.Def.Type == e.TypeAlly {
			out = append(out, c)
		}
	}
	return out
}

// drones lists the Drone minions engaged with you, facedown or not.
func drones(g *e.Game) []*e.Card {
	return minionsWithTrait(g, constants.TraitDrone)
}

// minionsWithTrait lists the engaged minions with the trait.
func minionsWithTrait(g *e.Game, trait string) []*e.Card {
	var out []*e.Card
	for _, m := range g.S.Minions {
		if m.Face().HasTrait(trait) {
			out = append(out, m)
		}
	}
	return out
}

// friendly lists your hero and allies.
func friendly(g *e.Game) []*e.Card { return append([]*e.Card{g.S.Hero}, allies(g)...) }

func enemiesWithoutWebbedUp(g *e.Game) []*e.Card {
	var out []*e.Card
	for _, en := range g.AllEnemies() {
		if findCode(en.Attached, constants.CodeWebbedUp) == nil {
			out = append(out, en)
		}
	}
	return out
}

// damagedCharacters lists every character with damage to heal: the identity,
// its allies, the villain and the minions.
func damagedCharacters(g *e.Game) []*e.Card {
	chars := []*e.Card{g.S.Hero}
	for _, c := range g.S.Play {
		if c.Def.Type == e.TypeAlly {
			chars = append(chars, c)
		}
	}
	var out []*e.Card
	for _, c := range append(chars, g.AllEnemies()...) {
		if c.Damage > 0 {
			out = append(out, c)
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

package engine

import (
	"slices"

	"mcbench/constants"
)

// Hero stats include bonuses from cards you control.

// heroCards are the cards whose bonuses apply to your hero.
func (g *Game) heroCards() []*Card {
	return append(append([]*Card(nil), g.S.Play...), g.S.Hero.Attached...)
}

func (g *Game) heroBonus(stat string, f func(*CardDef) int) (n int) {
	for _, c := range g.heroCards() {
		n += f(c.Def) + g.statBonus(c, stat)
	}
	return
}

func (g *Game) HeroATK() int {
	h := g.S.Hero
	return h.Face().ATK + h.TempATK + g.heroBonus(constants.StatHeroATK, func(d *CardDef) int { return d.HeroATK })
}

func (g *Game) HeroTHW() int {
	h := g.S.Hero
	return h.Face().THW + h.TempTHW + g.heroBonus(constants.StatHeroTHW, func(d *CardDef) int { return d.HeroTHW })
}

func (g *Game) HeroDEF() int {
	return g.S.Hero.Face().DEF + g.heroBonus(constants.StatHeroDEF, func(d *CardDef) int { return d.HeroDEF })
}

// HandSize is your current hand size, including your identity's modifiers.
func (g *Game) HandSize() int {
	return g.S.Hero.Face().HandSize + g.statBonus(g.S.Hero, constants.StatHandSize)
}

// HeroHasTrait reports whether your hero has the trait: printed, gained from
// a card you control, or gained until the end of the phase.
func (g *Game) HeroHasTrait(t string) bool {
	if g.S.Hero.Face().HasTrait(t) || slices.Contains(g.S.PhaseTraits, t) {
		return true
	}
	for _, c := range g.heroCards() {
		if slices.Contains(c.Def.HeroTraits, t) {
			return true
		}
	}
	return false
}

// HeroTraits lists your hero's current traits.
func (g *Game) HeroTraits() []string {
	out := append([]string(nil), g.S.Hero.Face().Traits...)
	for _, c := range g.heroCards() {
		out = append(out, c.Def.HeroTraits...)
	}
	out = append(out, g.S.PhaseTraits...)
	slices.Sort(out)
	return slices.Compact(out)
}

// AllyATK and AllyTHW include the ally's own scripted bonuses, its
// attachments and bonuses until the end of the phase.
func (g *Game) AllyATK(c *Card) int {
	return c.Def.ATK + g.statBonus(c, constants.StatATK) + g.attachedATK(c) + c.TempATK
}

func (g *Game) AllyTHW(c *Card) int {
	n := c.Def.THW + g.statBonus(c, constants.StatTHW) + c.TempTHW
	for _, a := range c.Attached {
		n += a.Face().AttachTHW
	}
	return n
}

func (g *Game) statBonus(c *Card, stat string) int {
	if sc := c.Face().Script; sc != nil && sc.StatBonus != nil {
		return sc.StatBonus(g, c, stat)
	}
	return 0
}

// MaxHP is a character's hit points with every "gets +X hit points" modifier.
func (g *Game) MaxHP(c *Card) int {
	s := g.S
	hp := c.Face().HP
	switch {
	case c == s.Hero:
		hp += g.heroBonus("", func(d *CardDef) int { return d.HeroHP })
	case c == s.Villain:
		for _, ss := range s.SideSchemes {
			hp += ss.Face().VillainHP
		}
	case c.Drone != nil:
		hp += g.droneBonus(func(d *CardDef) int { return d.DroneHP })
	}
	return hp
}

// RemainingHP is MaxHP less the damage on the character.
func (g *Game) RemainingHP(c *Card) int { return g.MaxHP(c) - c.Damage }

// droneBonus sums a modifier of the cards in play that boost facedown Drones.
func (g *Game) droneBonus(f func(*CardDef) int) (n int) {
	g.inPlay(func(c *Card) { n += f(c.Face()) })
	return
}

// AllyLimit is the number of allies you may control.
func (g *Game) AllyLimit() int {
	n := MaxAllies
	for _, c := range g.S.Play {
		n += c.Def.AllyLimit
	}
	return n
}

// endPhase ends the effects that last until the end of the phase.
func (g *Game) endPhase() {
	g.S.PhaseTraits = nil
	for _, c := range append([]*Card{g.S.Hero}, g.S.Play...) {
		c.TempATK, c.TempTHW = 0, 0
	}
}

func (g *Game) countInPlay(name string) (n int) {
	for _, c := range g.S.Play {
		if c.Def.Name == name {
			n++
		}
	}
	for _, c := range g.S.Hero.Attached {
		if c.Def.Name == name {
			n++
		}
	}
	return
}

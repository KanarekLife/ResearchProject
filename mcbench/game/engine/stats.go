package engine

// Stat names a StatBonus may be asked for.
const (
	statATK = "atk"
	statTHW = "thw"
)

// Hero stats include bonuses from cards you control.

func (g *Game) heroBonus(f func(*CardDef) int) (n int) {
	for _, c := range g.S.Play {
		n += f(c.Def)
	}
	for _, c := range g.S.Hero.Attached {
		n += f(c.Def)
	}
	return
}

func (g *Game) HeroATK() int {
	return g.S.Hero.Face().ATK + g.heroBonus(func(d *CardDef) int { return d.HeroATK })
}

func (g *Game) HeroTHW() int {
	return g.S.Hero.Face().THW + g.heroBonus(func(d *CardDef) int { return d.HeroTHW })
}

func (g *Game) HeroDEF() int {
	return g.S.Hero.Face().DEF + g.heroBonus(func(d *CardDef) int { return d.HeroDEF })
}

// AllyATK and AllyTHW include the ally's own scripted bonuses.
func (g *Game) AllyATK(c *Card) int { return c.Def.ATK + g.statBonus(c, statATK) }
func (g *Game) AllyTHW(c *Card) int { return c.Def.THW + g.statBonus(c, statTHW) }

func (g *Game) statBonus(c *Card, stat string) int {
	if sc := c.Def.Script; sc != nil && sc.StatBonus != nil {
		return sc.StatBonus(g, c, stat)
	}
	return 0
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

package engine

import (
	"fmt"

	"mcbench/constants"
)

// --- player attacks and thwarts --------------------------------------------------

// Attack performs an attack by a player character, honoring stun.
func (g *Game) Attack(attacker, target *Card, amount, consequential int) {
	if attacker.Stunned {
		attacker.Stunned = false
		g.Logf("%s is stunned: the attack is replaced by discarding the stun.", attacker.Name())
		return
	}
	g.Logf("%s attacks %s for %d.", attacker.Name(), target.Name(), amount)
	g.DamageEnemy(target, amount)
	if consequential > 0 && contains(g.S.Play, attacker) {
		g.DamageAlly(attacker, consequential)
	}
}

// Thwart performs a thwart by a player character, honoring confuse.
func (g *Game) Thwart(thwarter, scheme *Card, amount, consequential int) {
	if thwarter.Confused {
		thwarter.Confused = false
		g.Logf("%s is confused: the thwart is replaced by discarding the confuse.", thwarter.Name())
		return
	}
	g.Logf("%s thwarts %s for %d.", thwarter.Name(), scheme.Name(), amount)
	g.RemoveThreat(scheme, amount)
	if consequential > 0 && contains(g.S.Play, thwarter) {
		g.DamageAlly(thwarter, consequential)
	}
	g.fireForced(&Event{Trigger: TrigThwarted, Source: thwarter, Target: scheme})
}

// --- targeting --------------------------------------------------------------

// Enemies lists the enemies a player attack may target (guard enforced).
func (g *Game) Enemies() []*Card {
	var guards []*Card
	for _, m := range g.S.Minions {
		if m.Face().Guard {
			guards = append(guards, m)
		}
	}
	if len(guards) > 0 {
		return append([]*Card(nil), g.S.Minions...)
	}
	return append([]*Card{g.S.Villain}, g.S.Minions...)
}

// AllEnemies lists the villain and every minion, ignoring guard.
func (g *Game) AllEnemies() []*Card {
	return append([]*Card{g.S.Villain}, g.S.Minions...)
}

// ThwartableSchemes lists schemes with threat that may be thwarted.
func (g *Game) ThwartableSchemes() []*Card {
	var out []*Card
	if g.S.MainScheme.Threat > 0 && !g.CrisisActive() {
		out = append(out, g.S.MainScheme)
	}
	for _, s := range g.S.SideSchemes {
		if s.Threat > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (g *Game) HostOf(c *Card) *Card {
	var host *Card
	g.inPlay(func(x *Card) {
		if contains(x.Attached, c) {
			host = x
		}
	})
	return host
}

// ChooseScheme asks the player to pick a scheme with threat to remove n
// threat from (used by effects that say "remove N threat from a scheme").
func (g *Game) ChooseScheme(source string, n int) {
	g.Do(func() {
		schemes := g.ThwartableSchemes()
		if len(schemes) == 0 {
			return
		}
		var opts []Option
		for _, sc := range schemes {
			scheme := sc
			opts = append(opts, NewAction(
				constants.PrefixEffect+source+constants.SepTarget+g.Label(scheme),
				fmt.Sprintf("Remove %d threat from %s", n, g.Label(scheme)),
				func() { g.RemoveThreat(scheme, n) },
			))
		}
		g.Ask(&Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: choose a scheme to remove %d threat from.", source, n), Options: opts})
	})
}

// ChooseEnemy asks the player to pick an enemy (guard ignored) for an effect.
func (g *Game) ChooseEnemy(source, verb string, apply func(*Card)) {
	g.Do(func() {
		var opts []Option
		for _, e := range g.AllEnemies() {
			enemy := e
			opts = append(opts, NewAction(
				constants.PrefixEffect+source+constants.SepTarget+g.Label(enemy),
				fmt.Sprintf("%s %s", verb, g.Label(enemy)),
				func() { apply(enemy) },
			))
		}
		g.Ask(&Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: choose an enemy.", source), Options: opts})
	})
}

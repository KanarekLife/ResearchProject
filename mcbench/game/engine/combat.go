package engine

import (
	"fmt"

	"mcbench/constants"
)

// --- player attacks and thwarts --------------------------------------------------

// Attack performs an attack by a player character, honoring stun. It
// reports whether the attack was made (a stun replaces it).
func (g *Game) Attack(attacker, target *Card, amount, consequential int) bool {
	return g.attack(attacker, target, amount, consequential, false)
}

// OverkillAttack is an Attack whose excess damage on a defeated minion goes
// to the villain.
func (g *Game) OverkillAttack(attacker, target *Card, amount int) bool {
	return g.attack(attacker, target, amount, 0, true)
}

func (g *Game) attack(attacker, target *Card, amount, consequential int, overkill bool) bool {
	if attacker.Stunned {
		attacker.Stunned = false
		g.Logf("%s is stunned: the attack is replaced by discarding the stun.", attacker.Name())
		return false
	}
	g.Logf("%s attacks %s for %d.", attacker.Name(), target.Name(), amount)
	// Scheduled first so it runs after every ability the attack triggers.
	g.Do(func() { g.consequentialDamage(attacker, consequential) })
	defeated, excess := g.DamageEnemy(target, amount)
	if overkill && defeated && excess > 0 && target != g.S.Villain {
		g.Logf("Overkill: %d excess damage goes to %s.", excess, g.S.Villain.Name())
		g.DamageEnemy(g.S.Villain, excess)
	}
	g.Do(func() { g.afterAttack(attacker, target, defeated) })
	return true
}

// afterAttack resolves what follows a player attack: forced abilities, then
// retaliate, then optional responses.
func (g *Game) afterAttack(attacker, target *Card, defeated bool) {
	ev := &Event{Trigger: TrigAttacked, Source: attacker, Target: target, Defeated: defeated}
	g.fireForced(ev)
	if n := g.retaliate(target); n > 0 && g.isInPlay(target) && !defeated {
		g.Logf("%s retaliates: %d damage to %s.", target.Name(), n, attacker.Name())
		if attacker == g.S.Hero {
			g.HarmHero(n, nil)
		} else if contains(g.S.Play, attacker) {
			g.DamageAlly(attacker, n)
		}
	}
	g.Do(func() { g.playerWindow(ev, fmt.Sprintf("%s attacked %s.", attacker.Name(), target.Name()), func() {}) })
}

// retaliate is the character's retaliate value, including its attachments.
func (g *Game) retaliate(c *Card) int {
	n := c.Face().Retaliate
	for _, a := range c.Attached {
		n += a.Face().Retaliate
	}
	return n
}

// Thwart performs a thwart by a player character, honoring confuse. It
// reports whether the thwart was made (a confuse replaces it).
func (g *Game) Thwart(thwarter, scheme *Card, amount, consequential int) bool {
	if thwarter.Confused {
		thwarter.Confused = false
		g.Logf("%s is confused: the thwart is replaced by discarding the confuse.", thwarter.Name())
		return false
	}
	g.Logf("%s thwarts %s for %d.", thwarter.Name(), scheme.Name(), amount)
	// Scheduled first so it runs after every ability the thwart triggers.
	g.Do(func() { g.consequentialDamage(thwarter, consequential) })
	g.RemoveThreat(scheme, amount)
	ev := &Event{Trigger: TrigThwarted, Source: thwarter, Target: scheme}
	g.fireForced(ev)
	g.Do(func() { g.playerWindow(ev, fmt.Sprintf("%s thwarted %s.", thwarter.Name(), scheme.Name()), func() {}) })
	return true
}

// ThwartEach is one thwart for amount against every scheme with threat,
// honoring confuse. It reports whether the thwart was made.
func (g *Game) ThwartEach(thwarter *Card, amount int) bool {
	if thwarter.Confused {
		thwarter.Confused = false
		g.Logf("%s is confused: the thwart is replaced by discarding the confuse.", thwarter.Name())
		return false
	}
	for _, scheme := range g.ThwartableSchemes() {
		g.Logf("%s thwarts %s for %d.", thwarter.Name(), scheme.Name(), amount)
		g.RemoveThreat(scheme, amount)
		g.fireForced(&Event{Trigger: TrigThwarted, Source: thwarter, Target: scheme})
	}
	return true
}

func (g *Game) consequentialDamage(ally *Card, n int) {
	if n > 0 && contains(g.S.Play, ally) {
		g.DamageAlly(ally, n)
	}
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
	if g.S.MainScheme.Threat > 0 && !g.CrisisActive() && !g.S.MainScheme.Face().NoThreatRemoval {
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
// except, if set, is a scheme that may not be picked.
func (g *Game) ChooseScheme(source string, n int, except *Card) {
	g.Do(func() {
		var opts []Option
		for _, sc := range g.ThwartableSchemes() {
			if sc == except {
				continue
			}
			scheme := sc
			opts = append(opts, NewAction(
				constants.PrefixEffect+source+constants.SepTarget+g.Label(scheme),
				fmt.Sprintf("Remove %d threat from %s", n, g.Label(scheme)),
				func() { g.RemoveThreat(scheme, n) },
			))
		}
		if len(opts) == 0 {
			return
		}
		g.Ask(&Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: choose a scheme to remove %d threat from.", source, n), Options: opts})
	})
}

// ChooseEnemy asks the player to pick an enemy (guard ignored) for an effect.
func (g *Game) ChooseEnemy(source, verb string, apply func(*Card)) {
	g.ChooseCard(source, verb, "an enemy", func() []*Card { return g.AllEnemies() }, apply)
}

// ChooseCard asks the player to pick one of the cards listed when the step
// runs, for an effect described by verb ("Deal 2 damage to"); what names the
// kind of card in the prompt ("an enemy"). Nothing happens when there is none.
func (g *Game) ChooseCard(source, verb, what string, cards func() []*Card, apply func(*Card)) {
	g.Do(func() {
		var opts []Option
		for _, c := range cards() {
			card := c
			opts = append(opts, NewAction(
				constants.PrefixEffect+source+constants.SepTarget+g.Label(card),
				fmt.Sprintf("%s %s", verb, g.Label(card)),
				func() { apply(card) },
			))
		}
		if len(opts) == 0 {
			return
		}
		g.Ask(&Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: choose %s.", source, what), Options: opts})
	})
}

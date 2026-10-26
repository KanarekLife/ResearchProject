package cards

import (
	"fmt"
	"slices"

	"mcbench/constants"
	e "mcbench/game/engine"
)

// run executes a list of effects, skipping those whose conditions fail.
func run(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, effects []Effect) {
	for i := range effects {
		eff := effects[i]
		if !when(g, c, ev, eff.When) {
			continue
		}
		runEffect(g, c, ev, target, &eff)
	}
}

// runEffect applies one effect by trying each verb group in turn.
func runEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	for _, group := range []func(*e.Game, *e.Card, *e.Event, *e.Card, *Effect) bool{
		damageVerbs, threatVerbs, statusVerbs, zoneVerbs, timingVerbs, miscVerbs,
		playerCardVerbs, encounterVerbs,
	} {
		if group(g, c, ev, target, eff) {
			return
		}
	}
	panic("cards: unknown effect verb " + eff.Verb)
}

func damageVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbAttack:
		attackEffect(g, c, ev, target, eff, amount(g, c, eff.Damage, eff))
	case constants.VerbDamageEnemy:
		damageEnemyEffect(g, c, ev, target, eff)
	case constants.VerbDamageHero:
		n := amount(g, c, eff.Amount, eff)
		if eff.Result == constants.ResultEventAmount && ev != nil {
			n = ev.Amount
		}
		g.HarmHero(n, nil)
	case constants.VerbHeal:
		if eff.Target != "" {
			target = effectTarget(g, c, ev, target, eff)
		}
		g.Heal(target, eff.Amount)
	case constants.VerbHealHero:
		g.HealHero(eff.Amount)
	case constants.VerbHealVillain:
		if g.HealVillain(amount(g, c, eff.Amount, eff)) == 0 {
			run(g, c, ev, target, eff.IfZero)
		}
	default:
		return false
	}
	return true
}

func threatVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbThwart:
		thwartEffect(g, c, ev, target, eff)
	case constants.VerbThwartEach:
		g.ThwartEach(g.S.Hero, eff.Amount)
	case constants.VerbRemoveThreat:
		var except *e.Card
		if eff.Other {
			except = target
		}
		if eff.Choose || target == nil {
			g.ChooseScheme(c.Def.Name, eff.Amount, except)
		} else {
			g.RemoveThreat(target, eff.Amount)
		}
	case constants.VerbPlaceThreat:
		placeThreatEffect(g, c, eff)
	case constants.VerbScheme:
		var after func(int)
		if eff.Result == constants.ResultMill {
			after = func(placed int) { mill(g, c, placed) }
		}
		g.Do(func() { g.EnemyScheme(g.S.Villain, true, after) })
	default:
		return false
	}
	return true
}

func statusVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbStun:
		if t := effectTarget(g, c, ev, target, eff); t != nil {
			g.Stun(t)
		}
	case constants.VerbStunChosen:
		g.ChooseEnemy(c.Def.Name, "Stun", g.Stun)
	case constants.VerbConfuse:
		confuse(g, c, ev, target, eff)
	case constants.VerbGiveTough:
		giveTough(g, c, ev, target, eff)
	case constants.VerbExhaust:
		effectTarget(g, c, ev, target, eff).Exhausted = true
	case constants.VerbReady:
		effectTarget(g, c, ev, target, eff).Exhausted = false
	case constants.VerbCounter:
		counterEffect(g, c, ev, target, eff)
	default:
		return false
	}
	return true
}

func zoneVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbDraw:
		g.Draw(eff.N)
	case constants.VerbDiscardRandom:
		if x := g.RandomHandCard(); x != nil {
			g.DiscardFromHand(x)
		}
	case constants.VerbDiscardRandomPlace:
		discardRandomPlaceThreat(g)
	case constants.VerbShuffleEncounter:
		g.Shuffle(g.S.EncDeck)
	case constants.VerbToEncounterDiscard:
		g.S.EncDiscard = append(g.S.EncDiscard, c)
	case constants.VerbRemoveFromGame:
		g.RemoveFromGame(effectTarget(g, c, ev, target, eff))
	case constants.VerbReveal:
		g.Reveal(effectTarget(g, c, ev, target, eff))
	case constants.VerbDetach:
		g.Detach(effectTarget(g, c, ev, target, eff))
	case constants.VerbMillKeep:
		millKeep(g, c, eff.N, resource(eff.Resource))
	case constants.VerbFindAndReveal:
		findAndReveal(g, eff.Code)
	case constants.VerbTakeRandomCard:
		takeRandomCard(g, c)
	case constants.VerbReturnAttached:
		returnAttached(g, c)
	case constants.VerbNemesis:
		nemesisReveal(g)
	case constants.VerbDiscardChosen:
		discardChosen(g, c, ev, target, eff)
	default:
		return false
	}
	return true
}

func timingVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbSurge:
		g.Surge()
	case constants.VerbCancel:
		if ev != nil {
			ev.Cancelled = true
		}
	case constants.VerbPreventDamage:
		if ev != nil {
			ev.Amount = 0
		}
	case constants.VerbReduceAmount:
		if ev != nil {
			ev.Amount -= eff.Amount
		}
	case constants.VerbTakeThreatAsDamage:
		takeThreatAsDamage(g, c, ev)
	case constants.VerbAbsorbDamage:
		absorbDamage(g, c, ev, eff.Max)
	default:
		return false
	}
	return true
}

func miscVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbFlipAlterEgo:
		if g.IsHero() {
			from := g.S.Hero.Name()
			g.S.Hero.Flipped = true
			g.Logf("%s changes form to %s.", from, g.S.Hero.Name())
		}
	case constants.VerbVillainAttack:
		villainAttackEffect(g, c, eff)
	case constants.VerbVillainAndMinions:
		villainAndMinionsAttack(g)
	case constants.VerbCostReduction:
		g.S.CostReduction += eff.N
	case constants.VerbAssignDamage:
		assignBombDamage(g, c)
	case constants.VerbChoose:
		chooseEffect(g, c, ev, target, eff)
	default:
		return false
	}
	return true
}

// damageEnemyEffect deals damage to the effect's target, or asks the player
// to choose one when the effect is marked choose.
func damageEnemyEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	n := amount(g, c, eff.Amount, eff)
	if !eff.Choose {
		if eff.Target != "" {
			target = effectTarget(g, c, ev, target, eff)
		}
		g.DamageEnemy(target, n)
		return
	}
	g.ChooseEnemy(c.Def.Name, fmt.Sprintf("Deal %d damage to", n),
		func(t *e.Card) { g.DamageEnemy(t, n) })
}

// amount is base multiplied by the effect's per count (if any), capped by
// its max (if any).
func amount(g *e.Game, c *e.Card, base int, eff *Effect) int {
	if eff.Per != "" {
		base *= evalCount(g, c, eff.Per)
	}
	if eff.Max > 0 {
		base = min(base, eff.Max)
	}
	return base
}

// attackEffect makes your hero attack for n: the target, or one the player
// chooses when the effect is marked choose. Then runs if the attack is made.
func attackEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect, n int) {
	hit := func(t *e.Card) {
		var made bool
		if eff.Overkill {
			made = g.OverkillAttack(g.S.Hero, t, n)
		} else {
			made = g.Attack(g.S.Hero, t, n, 0)
		}
		if made {
			run(g, c, ev, t, eff.Then)
		}
	}
	if !eff.Choose {
		hit(target)
		return
	}
	g.ChooseCard(c.Def.Name, fmt.Sprintf("Attack for %d:", n), "an enemy to attack", g.Enemies, hit)
}

// thwartEffect makes your hero thwart the target, or a scheme the player
// chooses when the effect is marked choose; Then runs if the thwart is made.
// Without either it only removes threat from a chosen scheme.
func thwartEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	hit := func(t *e.Card) {
		if g.Thwart(g.S.Hero, t, eff.Amount, 0) {
			run(g, c, ev, t, eff.Then)
		}
	}
	switch {
	case eff.Choose:
		g.ChooseCard(c.Def.Name, fmt.Sprintf("Thwart for %d:", eff.Amount), "a scheme to thwart", g.ThwartableSchemes, hit)
	case target != nil:
		hit(target)
	default:
		g.ChooseScheme(c.Def.Name, eff.Amount, nil)
	}
}

// placeThreatEffect places threat on the main scheme, or on the effect's
// target: the card itself (self), each side scheme, or a side scheme by code.
func placeThreatEffect(g *e.Game, c *e.Card, eff *Effect) {
	n := amount(g, c, eff.Amount, eff)
	switch {
	case eff.Code != "":
		if s := findCode(g.S.SideSchemes, eff.Code); s != nil {
			g.PlaceThreatWindowed(s, n)
		}
	case eff.Target == constants.SelSelf:
		g.PlaceThreatWindowed(c, n)
	case eff.Target == constants.SelSideSchemes:
		for _, s := range append([]*e.Card(nil), g.S.SideSchemes...) {
			g.PlaceThreatWindowed(s, n)
		}
	default:
		g.PlaceThreatWindowed(g.S.MainScheme, n)
	}
}

// counterEffect changes a card's counters and detaches it when it empties.
func counterEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	t := effectTarget(g, c, ev, target, eff)
	t.Counters += eff.N
	if eff.DetachWhenEmpty && t.Counters <= 0 {
		g.Detach(t)
	}
}

// takeThreatAsDamage replaces incoming threat with damage to the hero.
func takeThreatAsDamage(g *e.Game, c *e.Card, ev *e.Event) {
	if ev == nil {
		return
	}
	n := ev.Amount
	ev.Amount = 0
	g.Logf("%s: %d threat is taken as damage instead.", c.Def.Name, n)
	g.HarmHero(n, nil)
}

// villainAttackEffect makes the villain attack; result: stun stuns whoever is
// damaged.
func villainAttackEffect(g *e.Game, c *e.Card, eff *Effect) {
	var onDamaged func(*e.Card, int)
	switch eff.Result {
	case constants.ResultStun:
		onDamaged = func(d *e.Card, _ int) {
			if d == g.S.Hero || slices.Contains(g.S.Play, d) {
				g.Stun(d)
			}
		}
	case constants.ResultThreat:
		placed := false
		onDamaged = func(*e.Card, int) {
			if !placed {
				placed = true
				g.PlaceThreatWindowed(g.S.MainScheme, 1)
			}
		}
	case constants.ResultMill:
		onDamaged = func(_ *e.Card, n int) { mill(g, c, n) }
	}
	g.Do(func() { g.EnemyAttack(g.S.Villain, true, onDamaged) })
}

func confuse(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	t := effectTarget(g, c, ev, target, eff)
	if t == nil {
		return
	}
	if t.Confused {
		run(g, c, ev, target, eff.IfAlready)
	} else {
		g.Confuse(t)
	}
}

func giveTough(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	t := effectTarget(g, c, ev, target, eff)
	if t == nil {
		return
	}
	if t.Tough {
		run(g, c, ev, target, eff.IfAlready)
	} else {
		g.GiveTough(t)
	}
}

func chooseEffect(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	options, prompt := eff.Options, eff.Prompt
	g.Do(func() {
		var opts []e.Option
		for _, ch := range options {
			if !when(g, c, ev, ch.When) {
				continue
			}
			ch := ch
			do := func() { run(g, c, ev, target, ch.Effects) }
			if ch.Paid != nil {
				again := func() { chooseEffect(g, c, ev, target, eff) }
				opts = append(opts, g.NewChoicePaid(nil, ch.Key, ch.Text, ch.Paid.Count, resources(ch.Paid.Resources), do, again)...)
				continue
			}
			opts = append(opts, e.NewAction(ch.Key, ch.Text, do))
		}
		if len(opts) == 0 {
			return
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: prompt, Options: opts})
	})
}

func discardChosen(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	sel, prompt, ifEmpty := eff.Target, eff.Prompt, eff.IfEmpty
	g.Do(func() {
		cards := selectTargets(g, c, sel)
		if len(cards) == 0 {
			run(g, c, ev, target, ifEmpty)
			return
		}
		var opts []e.Option
		for _, x := range cards {
			card := x
			opts = append(opts, e.NewAction(constants.PrefixEffect+c.Def.Name+">"+g.Label(card), "Discard "+g.Label(card), func() {
				g.Logf("%s is discarded.", card.Def.Name)
				g.Detach(card)
			}))
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: prompt, Options: opts})
	})
}

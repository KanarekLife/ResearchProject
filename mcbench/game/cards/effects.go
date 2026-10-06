package cards

import (
	"fmt"

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
		g.Attack(g.S.Hero, target, eff.Damage, 0)
	case constants.VerbDamageEnemy:
		damageEnemyEffect(g, c, target, eff)
	case constants.VerbDamageHero:
		amount := eff.Amount
		if eff.Result == constants.ResultEventAmount && ev != nil {
			amount = ev.Amount
		}
		g.DamageHero(amount)
	case constants.VerbHeal:
		g.Heal(target, eff.Amount)
	case constants.VerbHealHero:
		g.HealHero(eff.Amount)
	case constants.VerbHealVillain:
		if g.HealVillain(eff.Amount) == 0 {
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
		if target != nil {
			g.Thwart(g.S.Hero, target, eff.Amount, 0)
		} else {
			g.ChooseScheme(c.Def.Name, eff.Amount)
		}
	case constants.VerbRemoveThreat:
		if eff.Choose || target == nil {
			g.ChooseScheme(c.Def.Name, eff.Amount)
		} else {
			g.RemoveThreat(target, eff.Amount)
		}
	case constants.VerbPlaceThreat:
		g.PlaceThreatWindowed(g.S.MainScheme, eff.Amount)
	case constants.VerbScheme:
		g.Do(func() { g.EnemyScheme(g.S.Villain, true) })
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
		villainAttackEffect(g, eff)
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

// damageEnemyEffect deals damage to the current target, or asks the player
// to choose one when the effect is marked choose.
func damageEnemyEffect(g *e.Game, c, target *e.Card, eff *Effect) {
	if !eff.Choose {
		g.DamageEnemy(target, eff.Amount)
		return
	}
	g.ChooseEnemy(c.Def.Name, fmt.Sprintf("Deal %d damage to", eff.Amount),
		func(t *e.Card) { g.DamageEnemy(t, eff.Amount) })
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
	g.DamageHero(n)
}

// villainAttackEffect makes the villain attack; result: stun stuns whoever is
// damaged.
func villainAttackEffect(g *e.Game, eff *Effect) {
	var onDamaged func(*e.Card)
	if eff.Result == constants.ResultStun {
		onDamaged = g.Stun
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
			opts = append(opts, e.NewAction(ch.Key, ch.Text, func() { run(g, c, ev, target, ch.Effects) }))
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
		cards := selectTargets(g, sel)
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

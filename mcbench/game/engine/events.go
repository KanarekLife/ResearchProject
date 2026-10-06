package engine

import (
	"fmt"

	"mcbench/constants"
)

// Trigger names a timing point that abilities can react to.
type Trigger string

const (
	// An enemy is about to attack (before boost). Cancel to stop the attack.
	TrigEnemyWouldAttack Trigger = constants.TrigEnemyWouldAttack
	// The villain initiated an attack against the hero.
	TrigVillainAttacks Trigger = constants.TrigVillainAttacks
	// The hero would take damage from an attack. Amount may be reduced.
	TrigWouldTakeAttackDamage Trigger = constants.TrigWouldTakeAttackDamage
	// Damage would be dealt to an enemy. Amount may be redirected.
	TrigEnemyWouldTakeDamage Trigger = constants.TrigEnemyWouldTakeDamage
	// A treachery was revealed. Cancel to skip its When Revealed effect.
	TrigTreacheryRevealed Trigger = constants.TrigTreacheryRevealed
	// A minion is about to be defeated.
	TrigMinionDefeated Trigger = constants.TrigMinionDefeated
	// The villain is about to place threat by scheming. Amount may be reduced.
	TrigVillainSchemes Trigger = constants.TrigVillainSchemes
	// Threat is about to be placed on a scheme in the villain phase.
	TrigThreatWouldBePlaced Trigger = constants.TrigThreatWouldBePlaced
	// A player character finished thwarting (Source is the thwarter).
	TrigThwarted Trigger = constants.TrigThwarted
	// The round is ending.
	TrigRoundEnd Trigger = constants.TrigRoundEnd
	// An enemy attack finished resolving (Source is the enemy).
	TrigAttackEnded Trigger = constants.TrigAttackEnded
)

// Event carries the details of a trigger.
type Event struct {
	Trigger   Trigger
	Source    *Card // e.g. the attacker
	Target    *Card // e.g. the damaged card
	Amount    int
	Cancelled bool
}

// fireForced runs forced abilities of every card in play for this event.
func (g *Game) fireForced(ev *Event) {
	var hooks []func()
	g.inPlay(func(c *Card) {
		sc := c.Face().Script
		if sc == nil || sc.Forced == nil {
			return
		}
		if h := sc.Forced[ev.Trigger]; h != nil {
			card := c
			hooks = append(hooks, func() { h(g, card, ev) })
		}
	})
	// Collect first, then run, so hooks may change zones safely.
	for _, h := range hooks {
		h()
	}
}

// playerWindow offers interrupt/response events from hand for the trigger,
// then continues with next. At most one card is played per window.
func (g *Game) playerWindow(ev *Event, prompt string, next func()) {
	var opts []Option
	for _, c := range g.S.Hand {
		sc := c.Def.Script
		if sc == nil || sc.PlayWindow != ev.Trigger || sc.OnPlayEvent == nil {
			continue
		}
		if sc.PlayIf != nil && !sc.PlayIf(g, c, ev) {
			continue
		}
		card := c
		for _, pay := range g.paymentsFor(card, card.Def, max(0, card.Def.Cost-g.S.CostReduction), nil) {
			p := pay
			opts = append(opts, Option{
				Key:  constants.PrefixPlay + card.Name(),
				Pay:  p.Label(),
				Text: fmt.Sprintf("Play %s%s", card.Name(), p.Describe()),
				do: func() {
					p.Spend(g)
					g.S.CostReduction = 0
					g.moveFromHand(card)
					g.Logf("Played %s.", card.Name())
					card.Def.Script.OnPlayEvent(g, card, ev)
					g.S.Discard = append(g.S.Discard, card)
					g.Do(next)
				},
			})
		}
	}
	if len(opts) == 0 {
		g.Do(next)
		return
	}
	opts = append(opts, Option{Key: constants.KeyPass, Text: "Do not respond", do: func() { g.Do(next) }})
	g.Ask(&Decision{Kind: constants.KindWindow, Prompt: prompt, Options: opts})
}

package engine

import "mcbench/constants"

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
	// Threat is about to be placed on a scheme. Amount may be reduced.
	TrigThreatWouldBePlaced Trigger = constants.TrigThreatWouldBePlaced
	// A player character finished thwarting (Source is the thwarter).
	TrigThwarted Trigger = constants.TrigThwarted
	// The round is ending.
	TrigRoundEnd Trigger = constants.TrigRoundEnd
	// An enemy attack finished resolving (Source is the enemy).
	TrigAttackEnded Trigger = constants.TrigAttackEnded
	// A player character attacked (Source) an enemy (Target); Defeated tells
	// whether the attack defeated it.
	TrigAttacked Trigger = constants.TrigAttacked
	// An enemy attack (Source) damaged a character (Target) by Amount.
	TrigAttackDamaged Trigger = constants.TrigAttackDamaged
	// A player card (Target) entered play under your control.
	TrigEntersPlay Trigger = constants.TrigEntersPlay
	// A minion (Target) entered play engaged with you.
	TrigMinionEnters Trigger = constants.TrigMinionEnters
	// An enemy (Target) took damage and survived.
	TrigEnemyDamaged Trigger = constants.TrigEnemyDamaged
	// The villain (Source) finished scheming.
	TrigVillainSchemed Trigger = constants.TrigVillainSchemed
	// Threat was placed on the main scheme in step one of the villain phase.
	TrigStepOneThreat Trigger = constants.TrigStepOneThreat
	// Your hero would take damage (from any source). Amount may be reduced.
	// Abilities for it are also offered when the damage comes from an attack.
	TrigHeroWouldTakeDamage Trigger = constants.TrigHeroWouldTakeDamage
)

// Event carries the details of a trigger.
type Event struct {
	Trigger   Trigger
	Source    *Card // e.g. the attacker
	Target    *Card // e.g. the damaged card
	Amount    int
	Cancelled bool
	Defeated  bool // TrigAttacked: the attack defeated its target
	Attack    bool // OnBoost: the boost card is for an attack, not a scheme
}

// matches reports whether an ability waiting for trigger t answers ev.
func (ev *Event) matches(t Trigger) bool {
	return t == ev.Trigger || (t == TrigHeroWouldTakeDamage && ev.Trigger == TrigWouldTakeAttackDamage)
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

// playerWindow offers interrupt/response events from hand and abilities of
// cards in play for the trigger, then continues with next. At most one card
// or ability is used per window.
func (g *Game) playerWindow(ev *Event, prompt string, next func()) {
	opts := g.windowOptions(ev, prompt, next)
	if len(opts) == 0 {
		g.Do(next)
		return
	}
	opts = append(opts, Option{Key: constants.KeyPass, Text: "Do not respond", do: func() { g.Do(next) }})
	g.Ask(&Decision{Kind: constants.KindWindow, Prompt: prompt, Options: opts})
}

// windowOptions lists what may be used in the window, without the pass option.
func (g *Game) windowOptions(ev *Event, prompt string, next func()) []Option {
	var opts []Option
	seen := map[string]bool{}
	for _, c := range g.S.Hand {
		sc := c.Def.Script
		if sc == nil || !ev.matches(sc.PlayWindow) || sc.OnPlayEvent == nil {
			continue
		}
		if sc.PlayIf != nil && !sc.PlayIf(g, c, ev) {
			continue
		}
		card := c
		pays := g.paymentsFor(card, card.Def, max(0, card.Def.Cost-g.S.CostReduction), nil)
		if len(pays) == 0 {
			continue
		}
		play := func(p Payment) {
			p.Spend(g)
			g.S.CostReduction = 0
			g.moveFromHand(card)
			g.Logf("Played %s.", card.Name())
			// next is scheduled first so steps the event schedules run before it.
			g.Do(next)
			card.Def.Script.OnPlayEvent(g, card, ev)
			g.S.Discard = append(g.S.Discard, card)
		}
		// Cancelling a payment offers the same window again.
		back := func() { g.playerWindow(ev, prompt, next) }
		o := g.payOption(constants.PrefixPlay+card.Name(), "Play "+card.Name(), pays, play, back)
		if seen[o.FullKey()] {
			continue
		}
		seen[o.FullKey()] = true
		opts = append(opts, o)
	}
	g.inPlay(func(c *Card) {
		sc := c.Face().Script
		if sc == nil {
			return
		}
		for _, w := range sc.Windows {
			if !ev.matches(w.Trigger) || (w.If != nil && !w.If(g, c, ev)) {
				continue
			}
			card, w := c, w
			opts = append(opts, Option{Key: w.Key, Text: w.Text, do: func() {
				g.Do(next)
				w.Do(g, card, ev)
			}})
		}
	})
	return opts
}

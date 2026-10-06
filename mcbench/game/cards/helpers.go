package cards

import (
	"fmt"
	"slices"

	"mcbench/constants"
	e "mcbench/game/engine"
)

func absorbDamage(g *e.Game, c *e.Card, ev *e.Event, maxDamage int) {
	if ev == nil || g.HostOf(c) != ev.Target || ev.Amount <= 0 {
		return
	}
	c.Damage += ev.Amount
	g.Logf("%s absorbs %d damage (%d on it).", c.Def.Name, ev.Amount, c.Damage)
	ev.Amount = 0
	if c.Damage >= maxDamage {
		g.Logf("%s is discarded.", c.Def.Name)
		c.Damage = 0
		g.Detach(c)
	}
}

func millKeep(g *e.Game, c *e.Card, n int, keep e.Resource) {
	for i := 0; i < n && len(g.S.Deck) > 0; i++ {
		top := g.S.Deck[0]
		g.S.Deck = g.S.Deck[1:]
		if slices.Contains(top.Def.Resources, keep) {
			g.S.Hand = append(g.S.Hand, top)
			g.Logf("%s finds %s (added to hand).", c.Def.Name, top.Def.Name)
		} else {
			g.S.Discard = append(g.S.Discard, top)
			g.Logf("%s discards %s.", c.Def.Name, top.Def.Name)
		}
	}
}

func findAndReveal(g *e.Game, code string) {
	for _, zone := range []*[]*e.Card{&g.S.EncDeck, &g.S.EncDiscard} {
		for _, x := range *zone {
			if x.Def.Code == code {
				*zone = removeCard(*zone, x)
				g.Shuffle(g.S.EncDeck)
				card := x
				g.Do(func() { g.Reveal(card) })
				return
			}
		}
	}
	g.Shuffle(g.S.EncDeck)
}

func takeRandomCard(g *e.Game, c *e.Card) {
	x := g.RandomHandCard()
	if x == nil {
		return
	}
	g.S.Hand = removeCard(g.S.Hand, x)
	c.Attached = append(c.Attached, x)
	g.Logf("%s takes %s from your hand.", c.Def.Name, x.Def.Name)
}

func returnAttached(g *e.Game, c *e.Card) {
	for _, x := range c.Attached {
		g.S.Hand = append(g.S.Hand, x)
		g.Logf("%s returns to your hand.", x.Def.Name)
	}
	c.Attached = nil
}

func discardRandomPlaceThreat(g *e.Game) {
	x := g.RandomHandCard()
	if x == nil {
		return
	}
	g.DiscardFromHand(x)
	kinds := slices.Compact(slices.Sorted(slices.Values(x.Def.Resources)))
	g.PlaceThreat(g.S.MainScheme, len(kinds))
}

func villainAndMinionsAttack(g *e.Game) {
	steps := []func(){func() { g.EnemyAttack(g.S.Villain, true, nil) }}
	for _, m := range append([]*e.Card(nil), g.S.Minions...) {
		minion := m
		steps = append(steps, func() {
			if slices.Contains(g.S.Minions, minion) {
				g.EnemyAttack(minion, false, nil)
			}
		})
	}
	g.Do(steps...)
}

func assignBombDamage(g *e.Game, c *e.Card) {
	bomb := findCode(g.S.SideSchemes, constants.CodeBombScare)
	if bomb == nil {
		return
	}
	g.Logf("%s: %d damage to assign.", c.Def.Name, bomb.Threat)
	assignDamage(g, c.Def.Name, bomb.Threat)
}

// nemesisReveal resolves Shadow of the Past: the nemesis minion and side
// scheme enter play, the rest of the set is shuffled into the encounter deck,
// and it surges if no minion arrived.
func nemesisReveal(g *e.Game) {
	s := g.S
	minionEntered := false
	var steps []func()
	for _, n := range s.Nemesis {
		card := n
		switch card.Def.Type {
		case e.TypeMinion:
			minionEntered = true
			steps = append(steps, func() { g.EngageMinion(card) })
		case e.TypeSideScheme:
			steps = append(steps, func() { g.Reveal(card) })
		default:
			s.EncDeck = append(s.EncDeck, card)
		}
	}
	s.Nemesis = nil
	g.Shuffle(s.EncDeck)
	if len(steps) > 0 {
		g.Logf("Shadow of the Past: your nemesis arrives.")
	}
	if !minionEntered {
		steps = append(steps, g.Surge)
	}
	g.Do(steps...)
}

// assignDamage lets the player split n damage among the hero and allies one
// point at a time (all to the hero automatically when there are no allies).
func assignDamage(g *e.Game, source string, n int) {
	if n <= 0 {
		return
	}
	var allies []*e.Card
	for _, c := range g.S.Play {
		if c.Def.Type == e.TypeAlly {
			allies = append(allies, c)
		}
	}
	if len(allies) == 0 {
		g.DamageHero(n)
		return
	}
	g.Do(func() {
		opts := []e.Option{e.NewAction(constants.PrefixEffect+source+">"+g.S.Hero.Name(), "Deal 1 to "+g.S.Hero.Name(), func() {
			g.DamageHero(1)
			assignDamage(g, source, n-1)
		})}
		for _, a := range allies {
			ally := a
			opts = append(opts, e.NewAction(constants.PrefixEffect+source+">"+g.Label(ally), fmt.Sprintf("Deal 1 to %s (%d HP left)", g.Label(ally), ally.RemainingHP()), func() {
				g.DamageAlly(ally, 1)
				assignDamage(g, source, n-1)
			}))
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: assign 1 damage (%d left to assign).", source, n), Options: opts})
	})
}

func evalBonus(g *e.Game, expr string) int {
	if expr == constants.BonusSideSchemes {
		return len(g.S.SideSchemes)
	}
	panic("cards: unknown stat bonus " + expr)
}

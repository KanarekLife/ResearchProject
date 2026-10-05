package engine

import "fmt"

// MaxAllies is the ally limit per player.
const MaxAllies = 3

func (g *Game) playerTurn() {
	g.S.FormChanged = false
	g.Logf("--- Player turn, round %d ---", g.S.Round)
	g.phaseBegins(PhasePlayer)
	g.Do(g.askTurn)
}

// askTurn offers every legal action on the player's turn. Every action
// returns to askTurn until the player ends the turn.
func (g *Game) askTurn() {
	opts := g.TurnOptions()
	g.Ask(&Decision{Kind: "turn", Prompt: "Your turn. Choose an action, or end your turn.", Options: opts})
}

// then wraps an action so the turn continues afterwards.
func (g *Game) then(f func()) func() {
	return func() { g.Do(f, g.askTurn) }
}

// TurnOptions lists the legal actions in the player's main window.
func (g *Game) TurnOptions() []Option {
	s := g.S
	h := s.Hero
	var opts []Option

	// Play cards from hand.
	for _, c := range s.Hand {
		opts = append(opts, g.playOptions(c)...)
	}

	// Basic powers.
	if !h.Exhausted && g.IsHero() {
		for _, t := range g.Enemies() {
			target := t
			note := ""
			if h.Stunned {
				note = " (you are stunned: this only removes the stun)"
			}
			opts = append(opts, Option{
				Key:  "attack:" + h.Name() + ">" + g.Label(target),
				Text: fmt.Sprintf("Basic attack: %s deals %d damage to %s (exhaust %s)%s", h.Name(), g.HeroATK(), g.Label(target), h.Name(), note),
				do: g.then(func() {
					h.Exhausted = true
					g.Attack(h, target, g.HeroATK(), 0)
				}),
			})
		}
		for _, sc := range g.ThwartableSchemes() {
			scheme := sc
			note := ""
			if h.Confused {
				note = " (you are confused: this only removes the confuse)"
			}
			opts = append(opts, Option{
				Key:  "thwart:" + h.Name() + ">" + g.Label(scheme),
				Text: fmt.Sprintf("Basic thwart: %s removes %d threat from %s (exhaust %s)%s", h.Name(), g.HeroTHW(), g.Label(scheme), h.Name(), note),
				do: g.then(func() {
					h.Exhausted = true
					g.Thwart(h, scheme, g.HeroTHW(), 0)
				}),
			})
		}
	}
	if !h.Exhausted && !g.IsHero() && h.Damage > 0 {
		opts = append(opts, Option{
			Key:  "recover:" + h.Name(),
			Text: fmt.Sprintf("Recover: %s heals %d damage (exhaust %s)", h.Name(), h.Face().REC, h.Name()),
			do: g.then(func() {
				h.Exhausted = true
				g.HealHero(h.Face().REC)
			}),
		})
	}

	// Allies.
	for _, a := range s.Play {
		if a.Def.Type != TypeAlly || a.Exhausted {
			continue
		}
		ally := a
		for _, t := range g.Enemies() {
			target := t
			opts = append(opts, Option{
				Key:  "attack:" + g.Label(ally) + ">" + g.Label(target),
				Text: fmt.Sprintf("Ally attack: %s deals %d damage to %s (exhaust; takes %d consequential damage)", g.Label(ally), g.AllyATK(ally), g.Label(target), ally.Def.AtkCons),
				do: g.then(func() {
					ally.Exhausted = true
					g.Attack(ally, target, g.AllyATK(ally), ally.Def.AtkCons)
				}),
			})
		}
		for _, sc := range g.ThwartableSchemes() {
			scheme := sc
			opts = append(opts, Option{
				Key:  "thwart:" + g.Label(ally) + ">" + g.Label(scheme),
				Text: fmt.Sprintf("Ally thwart: %s removes %d threat from %s (exhaust; takes %d consequential damage)", g.Label(ally), g.AllyTHW(ally), g.Label(scheme), ally.Def.ThwCons),
				do: g.then(func() {
					ally.Exhausted = true
					g.Thwart(ally, scheme, g.AllyTHW(ally), ally.Def.ThwCons)
				}),
			})
		}
	}

	// Card abilities.
	g.inPlay(func(c *Card) {
		sc := c.Face().Script
		if sc == nil || sc.Actions == nil {
			return
		}
		for _, o := range sc.Actions(g, c) {
			o.do = g.then(o.do)
			opts = append(opts, o)
		}
	})

	if !s.FormChanged {
		other := h.Def.Name
		if !h.Flipped {
			other = h.Def.Back.Name
		}
		opts = append(opts, Option{
			Key:  "change_form:" + other,
			Text: fmt.Sprintf("Change form to %s (once per turn)", other),
			do: g.then(func() {
				from := h.Name()
				h.Flipped = !h.Flipped
				s.FormChanged = true
				g.Logf("%s changes form to %s.", from, h.Name())
			}),
		})
	}

	opts = append(opts, Option{Key: "end_turn", Text: "End your turn", do: func() { g.Do(g.endPlayerPhase) }})
	return opts
}

// NewAction builds an Option for a card's Actions hook. The turn continues
// automatically after it resolves.
func NewAction(key, text string, do func()) Option {
	return Option{Key: key, Text: text, do: do}
}

// NewActionPaid builds one Option per legal payment for a card ability.
func (g *Game) NewActionPaid(src *Card, key, text string, cost int, need []Resource, do func()) []Option {
	var out []Option
	for _, p := range g.paymentsFor(src, nil, cost, need) {
		pay := p
		out = append(out, Option{Key: key, Pay: pay.Label(), Text: text + pay.Describe(), do: func() {
			pay.Spend(g)
			do()
		}})
	}
	return out
}

func (g *Game) uniqueInPlay(name string) bool {
	found := false
	g.inPlay(func(c *Card) {
		if c.Name() == name {
			found = true
		}
	})
	return found
}

func (g *Game) allyCount() (n int) {
	for _, c := range g.S.Play {
		if c.Def.Type == TypeAlly {
			n++
		}
	}
	return
}

// playOptions lists the ways to play a card from hand on the player's turn.
func (g *Game) playOptions(c *Card) []Option {
	d := c.Def
	sc := d.Script
	switch d.Type {
	case TypeResource:
		return nil
	case TypeEvent:
		if sc == nil || sc.OnPlay == nil || sc.PlayWindow != "" {
			return nil
		}
	case TypeAlly:
		if g.allyCount() >= MaxAllies {
			return nil
		}
	}
	if d.Unique && g.uniqueInPlay(d.Name) {
		return nil
	}
	if d.MaxInPlay > 0 && g.countInPlay(d.Name) >= d.MaxInPlay {
		return nil
	}
	if sc != nil && sc.CanPlay != nil && !sc.CanPlay(g, c) {
		return nil
	}
	targets := []*Card{nil}
	if sc != nil && sc.Targets != nil {
		targets = sc.Targets(g, c)
		if len(targets) == 0 {
			return nil
		}
	}
	var opts []Option
	for _, t := range targets {
		target := t
		key := "play:" + d.Name
		text := "Play " + d.Name
		if target != nil {
			key += ">" + g.Label(target)
			text += " targeting " + g.Label(target)
		}
		for _, p := range g.paymentsFor(c, d, max(0, d.Cost-g.S.CostReduction), nil) {
			pay := p
			opts = append(opts, Option{
				Key:  key,
				Pay:  pay.Label(),
				Text: text + pay.Describe(),
				do: g.then(func() {
					pay.Spend(g)
					g.playCard(c, target)
				}),
			})
		}
	}
	return opts
}

func (g *Game) playCard(c *Card, target *Card) {
	s := g.S
	d := c.Def
	g.moveFromHand(c)
	s.CostReduction = 0
	g.Logf("Played %s.", d.Name)
	switch d.Type {
	case TypeEvent:
		d.Script.OnPlay(g, c, target)
		s.Discard = append(s.Discard, c)
		return
	case TypeUpgrade:
		if target != nil {
			target.Attached = append(target.Attached, c)
		} else {
			s.Play = append(s.Play, c)
		}
	default:
		s.Play = append(s.Play, c)
	}
	c.Counters = d.Uses
	if d.Script != nil && d.Script.OnPlay != nil {
		d.Script.OnPlay(g, c, target)
	}
}

// Draw draws n cards, reshuffling the discard pile when the deck runs out
// (which deals an encounter card face-down to the hero).
func (g *Game) Draw(n int) {
	s := g.S
	for i := 0; i < n; i++ {
		if len(s.Deck) == 0 {
			if len(s.Discard) == 0 {
				return
			}
			s.Deck, s.Discard = s.Discard, nil
			g.shuffle(s.Deck)
			g.Logf("Player deck reshuffled; an encounter card is dealt to %s.", s.Hero.Name())
			if c := g.drawEncounter(); c != nil {
				s.Dealt = append(s.Dealt, c)
			}
		}
		s.Hand = append(s.Hand, s.Deck[0])
		s.Deck = s.Deck[1:]
	}
	g.Logf("%s draws %d card(s).", s.Hero.Name(), n)
}

// endPlayerPhase: optional discards, draw to hand size, ready everything.
func (g *Game) endPlayerPhase() {
	g.Do(g.askDiscard)
}

func (g *Game) askDiscard() {
	s := g.S
	opts := []Option{{Key: "discard:done", Text: "Done discarding; draw up to hand size", do: func() { g.Do(g.drawAndReady) }}}
	seen := map[string]bool{}
	for _, c := range s.Hand {
		card := c
		if seen[card.Def.Name] {
			continue
		}
		seen[card.Def.Name] = true
		opts = append(opts, Option{
			Key:  "discard:" + card.Def.Name,
			Text: "Discard " + card.Def.Name + " from hand",
			do: func() {
				g.moveFromHand(card)
				s.Discard = append(s.Discard, card)
				g.Do(g.askDiscard)
			},
		})
	}
	if len(opts) == 1 {
		g.Do(g.drawAndReady)
		return
	}
	g.Ask(&Decision{Kind: "discard", Prompt: fmt.Sprintf("End of turn. You may discard cards from hand before drawing up to your hand size of %d.", s.Hero.Face().HandSize), Options: opts})
}

func (g *Game) drawAndReady() {
	s := g.S
	s.CostReduction = 0
	if need := s.Hero.Face().HandSize - len(s.Hand); need > 0 {
		g.Draw(need)
	}
	s.Hero.Exhausted = false
	for _, c := range s.Play {
		c.Exhausted = false
	}
	g.inPlay(func(c *Card) {
		if c.Def.IsPlayerCard() {
			c.Exhausted = false
		}
	})
	g.Do(g.villainPhase, g.endRound)
}

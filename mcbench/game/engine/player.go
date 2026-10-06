package engine

import (
	"fmt"

	"mcbench/constants"
)

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
	g.Ask(&Decision{Kind: constants.KindTurn, Prompt: "Your turn. Choose an action, or end your turn.", Options: g.TurnOptions()})
}

// then wraps an action so the turn continues afterwards.
func (g *Game) then(f func()) func() {
	return func() { g.Do(f, g.askTurn) }
}

func (g *Game) uniqueInPlay(name string) bool { return g.nameCount(name) > 0 }

func (g *Game) allyCount() (n int) {
	for _, c := range g.S.Play {
		if c.Def.Type == TypeAlly {
			n++
		}
	}
	return
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
	g.Do(func() { g.enforceAllyLimit(c) }, func() {
		// A card discarded for the ally limit never resolves its abilities.
		if d.Script != nil && d.Script.OnPlay != nil && g.isInPlay(c) {
			d.Script.OnPlay(g, c, target)
		}
	})
}

// enforceAllyLimit makes the player discard allies until they control no
// more than MaxAllies. It runs before the played card's enters-play abilities.
func (g *Game) enforceAllyLimit(played *Card) {
	n := g.allyCount()
	if n <= MaxAllies {
		return
	}
	var opts []Option
	for _, a := range g.S.Play {
		if a.Def.Type != TypeAlly {
			continue
		}
		ally := a
		label := g.Label(ally)
		opts = append(opts, NewAction(constants.PrefixEffect+played.Def.Name+constants.SepTarget+label, "Discard "+label, func() {
			g.Logf("%s is discarded (ally limit).", label)
			g.Detach(ally)
			g.Do(func() { g.enforceAllyLimit(played) })
		}))
	}
	g.Ask(&Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("Ally limit: you control %d allies but may control only %d. Choose an ally to discard.", n, MaxAllies), Options: opts})
}

// Draw draws n cards one at a time.
func (g *Game) Draw(n int) {
	s := g.S
	for i := 0; i < n; i++ {
		g.ResetEmptyDeck()
		if len(s.Deck) == 0 {
			return
		}
		s.Hand = append(s.Hand, s.Deck[0])
		s.Deck = s.Deck[1:]
	}
	g.ResetEmptyDeck()
	g.Logf("%s draws %d card(s).", s.Hero.Name(), n)
}

// ResetEmptyDeck shuffles the discard pile into a new deck as soon as the
// deck is empty, and deals the hero a face-down encounter card. With an empty
// discard pile the reset waits until a card is there.
func (g *Game) ResetEmptyDeck() {
	s := g.S
	if len(s.Deck) > 0 || len(s.Discard) == 0 {
		return
	}
	s.Deck, s.Discard = s.Discard, nil
	g.shuffle(s.Deck)
	g.Logf("Player deck reshuffled; an encounter card is dealt to %s.", s.Hero.Name())
	if c := g.drawEncounter(); c != nil {
		s.Dealt = append(s.Dealt, c)
	}
}

func (g *Game) endPlayerPhase() {
	g.Do(g.askDiscard)
}

// askDiscard lets the player discard from hand at the end of the turn. Over
// the hand size, the player must discard down to it before being done.
func (g *Game) askDiscard() {
	s := g.S
	handSize := s.Hero.Face().HandSize
	var opts []Option
	prompt := fmt.Sprintf("End of turn. You must discard down to your hand size of %d (you have %d cards).", handSize, len(s.Hand))
	if len(s.Hand) <= handSize {
		opts = append(opts, Option{Key: constants.KeyDiscardDone, Text: "Done discarding; draw up to hand size", do: func() { g.Do(g.drawAndReady) }})
		prompt = fmt.Sprintf("End of turn. You may discard cards from hand before drawing up to your hand size of %d.", handSize)
	}
	seen := map[string]bool{}
	for _, c := range s.Hand {
		card := c
		if seen[card.Def.Name] {
			continue
		}
		seen[card.Def.Name] = true
		opts = append(opts, Option{
			Key:  constants.PrefixDiscard + card.Def.Name,
			Text: "Discard " + card.Def.Name + " from hand",
			do: func() {
				g.moveFromHand(card)
				s.Discard = append(s.Discard, card)
				g.Do(g.askDiscard)
			},
		})
	}
	if len(s.Hand) == 0 {
		g.Do(g.drawAndReady)
		return
	}
	g.Ask(&Decision{Kind: constants.KindDiscard, Prompt: prompt, Options: opts})
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

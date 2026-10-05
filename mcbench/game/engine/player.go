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

func (g *Game) endPlayerPhase() {
	g.Do(g.askDiscard)
}

func (g *Game) askDiscard() {
	s := g.S
	opts := []Option{{Key: constants.KeyDiscardDone, Text: "Done discarding; draw up to hand size", do: func() { g.Do(g.drawAndReady) }}}
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
	if len(opts) == 1 {
		g.Do(g.drawAndReady)
		return
	}
	g.Ask(&Decision{Kind: constants.KindDiscard, Prompt: fmt.Sprintf("End of turn. You may discard cards from hand before drawing up to your hand size of %d.", s.Hero.Face().HandSize), Options: opts})
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

package engine

import (
	"fmt"
)

// drawEncounter takes the top encounter card. The deck is reset as soon as
// it is empty, so the card being drawn is not shuffled back in.
func (g *Game) drawEncounter() *Card {
	s := g.S
	g.resetEncounterDeck()
	if len(s.EncDeck) == 0 {
		return nil
	}
	c := s.EncDeck[0]
	s.EncDeck = s.EncDeck[1:]
	g.resetEncounterDeck()
	return c
}

// resetEncounterDeck shuffles the encounter discard pile into a new deck
// (with an acceleration token) as soon as the deck is empty. If the discard
// pile is empty too, the reset loops forever and the players lose.
func (g *Game) resetEncounterDeck() {
	s := g.S
	if len(s.EncDeck) > 0 {
		return
	}
	if len(s.EncDiscard) == 0 {
		g.lose("the encounter deck and its discard pile are both empty")
		return
	}
	s.EncDeck, s.EncDiscard = s.EncDiscard, nil
	g.shuffle(s.EncDeck)
	s.AccelTokens++
	g.Logf("Encounter deck reshuffled; an acceleration token is added (%d total).", s.AccelTokens)
}

// flipBoost flips a facedown boost card and discards it, returning its
// boost icons.
func (g *Game) flipBoost(c *Card) int {
	if c == nil {
		return 0
	}
	g.Logf("Boost card revealed: %s (+%d).", c.Def.Name, c.Def.Boost)
	g.S.EncDiscard = append(g.S.EncDiscard, c)
	return c.Def.Boost
}

// dealAndRevealEncounters deals 1 + hazard encounter cards, then reveals
// every dealt card one at a time, including cards dealt earlier in the round
// and cards dealt while revealing.
func (g *Game) dealAndRevealEncounters() {
	s := g.S
	n := 1
	for _, ss := range s.SideSchemes {
		n += ss.Face().Hazard
	}
	for i := 0; i < n; i++ {
		if c := g.drawEncounter(); c != nil {
			s.Dealt = append(s.Dealt, c)
		}
	}
	g.Do(g.revealDealt)
}

// revealDealt reveals the next dealt card, then the rest of the queue.
func (g *Game) revealDealt() {
	s := g.S
	if len(s.Dealt) == 0 {
		return
	}
	c := s.Dealt[0]
	s.Dealt = s.Dealt[1:]
	g.Do(func() { g.Reveal(c) }, g.revealDealt)
}

// Surge deals the player one more facedown encounter card. It joins the
// queue of dealt cards, revealed in the villain phase.
func (g *Game) Surge() {
	g.Logf("Surge! An encounter card is dealt facedown to %s.", g.S.Hero.Name())
	if c := g.drawEncounter(); c != nil {
		g.S.Dealt = append(g.S.Dealt, c)
	}
}

func (g *Game) Reveal(c *Card) {
	g.Logf("Encounter card revealed: %s (%s).", c.Def.Name, c.Def.Type)
	onReveal := func() {
		if d := c.Def; d.Script != nil && d.Script.OnReveal != nil {
			d.Script.OnReveal(g, c)
		}
	}
	steps := g.revealSteps(c, onReveal)
	if c.Def.Surge {
		steps = append(steps, g.Surge)
	}
	g.Do(steps...)
}

// revealSteps lists how a card enters play and resolves its When Revealed.
func (g *Game) revealSteps(c *Card, onReveal func()) []func() {
	s, d := g.S, c.Def
	switch d.Type {
	case TypeTreachery:
		ev := &Event{Trigger: TrigTreacheryRevealed, Source: c}
		return []func(){
			func() {
				g.playerWindow(ev, fmt.Sprintf("Treachery %s is being revealed: %s", d.Name, d.Text), func() {})
			},
			func() {
				if ev.Cancelled {
					g.Logf("The When Revealed effect of %s is cancelled.", d.Name)
				} else {
					onReveal()
				}
			},
			func() { s.EncDiscard = append(s.EncDiscard, c) },
		}
	case TypeMinion:
		return []func(){func() { g.EngageMinion(c); onReveal() }}
	case TypeSideScheme:
		return []func(){func() {
			c.Threat = d.StartingThreat
			s.SideSchemes = append(s.SideSchemes, c)
			g.Logf("Side scheme %s enters play with %d threat.", d.Name, c.Threat)
			onReveal()
		}}
	case TypeAttachment:
		return []func(){func() {
			s.Villain.Attached = append(s.Villain.Attached, c)
			g.Logf("%s attaches to %s.", d.Name, s.Villain.Name())
			onReveal()
		}}
	case TypeObligation:
		// The card's script decides where it goes (removed or discarded).
		return []func(){onReveal}
	default:
		return []func(){onReveal, func() { s.EncDiscard = append(s.EncDiscard, c) }}
	}
}

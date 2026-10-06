package engine

import (
	"fmt"
)

// drawEncounter takes the top encounter card, reshuffling the discard pile
// into the deck (with an acceleration token) when the deck is empty.
func (g *Game) drawEncounter() *Card {
	s := g.S
	if len(s.EncDeck) == 0 {
		if len(s.EncDiscard) == 0 {
			return nil
		}
		s.EncDeck, s.EncDiscard = s.EncDiscard, nil
		g.shuffle(s.EncDeck)
		s.AccelTokens++
		g.Logf("Encounter deck reshuffled; an acceleration token is added (%d total).", s.AccelTokens)
	}
	c := s.EncDeck[0]
	s.EncDeck = s.EncDeck[1:]
	return c
}

// boost deals and immediately resolves a boost card, returning its icons.
func (g *Game) boost() int {
	c := g.drawEncounter()
	if c == nil {
		return 0
	}
	g.Logf("Boost card revealed: %s (+%d).", c.Def.Name, c.Def.Boost)
	g.S.EncDiscard = append(g.S.EncDiscard, c)
	return c.Def.Boost
}

// dealAndRevealEncounters deals 1 + hazard encounter cards and reveals them,
// together with any cards dealt earlier in the round.
func (g *Game) dealAndRevealEncounters() {
	s := g.S
	n := 1
	for _, ss := range s.SideSchemes {
		n += ss.Face().Hazard
	}
	dealt := s.Dealt
	s.Dealt = nil
	for i := 0; i < n; i++ {
		if c := g.drawEncounter(); c != nil {
			dealt = append(dealt, c)
		}
	}
	var steps []func()
	for _, c := range dealt {
		card := c
		steps = append(steps, func() { g.Reveal(card) })
	}
	g.Do(steps...)
}

// Surge makes the current reveal reveal one more encounter card afterwards.
func (g *Game) Surge() {
	g.Logf("Surge!")
	g.Do(func() {
		if c := g.drawEncounter(); c != nil {
			g.Reveal(c)
		}
	})
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

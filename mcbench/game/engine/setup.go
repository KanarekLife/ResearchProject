package engine

import (
	"fmt"

	"mcbench/constants"
)

// StartGame performs solo game setup (Rules Reference order, simplified to
// one player) and runs until the first decision:
//   - the villain's stage I and the main scheme are in play (built by the caller),
//   - the encounter deck (with your obligation) and your deck are shuffled,
//   - you start in alter-ego form and draw up to your hand size,
//   - you may mulligan (discard any cards, then draw back up),
//   - round 1 begins with the player phase.
func (g *Game) StartGame() {
	s := g.S
	s.Round, s.Phase = 1, PhasePlayer
	s.Hero.Flipped = true
	s.MainScheme.Threat = s.MainScheme.Face().StartingThreat
	s.VillainTotalHP = s.Villain.Face().HP
	for _, st := range s.VillainStages {
		s.VillainTotalHP += st.HP
	}
	if s.Villain.Face().Toughness {
		s.Villain.Tough = true
	}
	g.shuffle(s.Deck)
	g.shuffle(s.EncDeck)
	g.Logf("Setup: %s vs %s (%s).", s.Hero.Def.Name, s.Villain.Name(), s.MainScheme.Name())
	g.Draw(s.Hero.Face().HandSize)
	g.Do(g.askMulligan, g.playerTurn)
	g.Run()
}

func (g *Game) askMulligan() {
	s := g.S
	opts := []Option{{Key: constants.KeyMulliganKeep, Text: "Keep this hand", do: func() {
		if need := s.Hero.Face().HandSize - len(s.Hand); need > 0 {
			g.Draw(need)
		}
	}}}
	seen := map[string]bool{}
	for _, c := range s.Hand {
		card := c
		if seen[card.Def.Name] {
			continue
		}
		seen[card.Def.Name] = true
		opts = append(opts, Option{
			Key:  constants.PrefixMulligan + card.Def.Name,
			Text: "Mulligan: discard " + card.Def.Name + " (you draw back up to your hand size when you keep)",
			do: func() {
				g.moveFromHand(card)
				s.Discard = append(s.Discard, card)
				g.Do(g.askMulligan)
			},
		})
	}
	g.Ask(&Decision{Kind: constants.KindMulligan, Prompt: fmt.Sprintf("Opening hand. Discard any cards you do not want; when you keep, you draw back up to %d.", s.Hero.Face().HandSize), Options: opts})
}

// EngageMinion puts a minion into play engaged with you. Quickstrike
// minions attack immediately if you are in hero form.
func (g *Game) EngageMinion(c *Card) {
	s := g.S
	s.Minions = append(s.Minions, c)
	g.Logf("%s engages %s.", c.Def.Name, s.Hero.Name())
	if c.Def.Toughness {
		g.GiveTough(c)
	}
	if c.Def.Quickstrike && g.IsHero() {
		g.Do(func() { g.EnemyAttack(c, false, nil) })
	}
}

// RandomHandCard picks a random card from hand (seeded), or nil.
func (g *Game) RandomHandCard() *Card {
	if len(g.S.Hand) == 0 {
		return nil
	}
	return g.S.Hand[g.rng.IntN(len(g.S.Hand))]
}

func (g *Game) DiscardFromHand(c *Card) {
	g.moveFromHand(c)
	g.S.Discard = append(g.S.Discard, c)
	g.Logf("%s is discarded from hand.", c.Def.Name)
}

func (g *Game) RemoveFromGame(c *Card) {
	g.S.Removed = append(g.S.Removed, c)
	g.Logf("%s is removed from the game.", c.Def.Name)
}

// BeginPlayerTurn starts at the player turn of the current round, skipping
// setup. It exists for rule unit tests that build a position by hand.
func (g *Game) BeginPlayerTurn() {
	if g.S.Round == 0 {
		g.S.Round = 1
	}
	g.S.Phase = PhasePlayer
	g.Do(g.playerTurn)
	g.Run()
}

package cards

import (
	"fmt"

	e "mcbench/engine"
)

// Standard encounter set, Bomb Scare modular set, and Spider-Man's
// obligation and nemesis set (Vulture).

func init() {
	// --- Standard ---------------------------------------------------------------
	add(&e.CardDef{
		Code: "01186", Name: "Advance", Type: e.TypeTreachery,
		Text: "The villain schemes.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			g.Do(func() { g.EnemyScheme(g.S.Villain, true) })
		}},
	})
	add(&e.CardDef{
		Code: "01187", Name: "Assault", Type: e.TypeTreachery,
		Text: "Alter-ego: surge. Hero: the villain attacks you.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if !g.IsHero() {
				g.Surge()
				return
			}
			g.Do(func() { g.EnemyAttack(g.S.Villain, true, nil) })
		}},
	})
	add(&e.CardDef{
		Code: "01188", Name: "Caught Off Guard", Type: e.TypeTreachery, Boost: 1,
		Text: "Discard an upgrade or support you control. If you cannot, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			var opts []e.Option
			for _, x := range controlledUpgradesSupports(g) {
				card := x
				opts = append(opts, e.NewAction("effect:Caught Off Guard>"+card.Def.Name, "Discard "+g.Label(card), func() {
					g.Logf("%s is discarded.", card.Def.Name)
					g.Detach(card)
				}))
			}
			if len(opts) == 0 {
				g.Surge()
				return
			}
			g.Do(func() {
				g.Ask(&e.Decision{Kind: "choice", Prompt: "Caught Off Guard: choose an upgrade or support you control to discard.", Options: opts})
			})
		}},
	})
	add(&e.CardDef{
		Code: "01189", Name: "Gang-Up", Type: e.TypeTreachery, Boost: 1,
		Text: "Alter-ego: surge. Hero: the villain and each engaged minion attack you.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if !g.IsHero() {
				g.Surge()
				return
			}
			steps := []func(){func() { g.EnemyAttack(g.S.Villain, true, nil) }}
			for _, m := range append([]*e.Card(nil), g.S.Minions...) {
				minion := m
				steps = append(steps, func() {
					if contains(g.S.Minions, minion) {
						g.EnemyAttack(minion, false, nil)
					}
				})
			}
			g.Do(steps...)
		}},
	})
	add(&e.CardDef{
		Code: "01190", Name: "Shadow of the Past", Type: e.TypeTreachery, Boost: 2,
		Text: "Your nemesis minion enters play engaged with you and your nemesis side scheme enters play; the rest of your nemesis set is shuffled into the encounter deck. If no nemesis minion entered play, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
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
		}},
	})

	// --- Bomb Scare (modular) ---------------------------------------------------
	add(&e.CardDef{
		Code: "01111", Name: "Explosion", Type: e.TypeTreachery, Boost: 2,
		Text: "If Bomb Scare is in play, deal X damage divided as you choose among your hero and allies (X = threat on Bomb Scare). Otherwise surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			var bomb *e.Card
			for _, s := range g.S.SideSchemes {
				if s.Def.Code == "01109" {
					bomb = s
				}
			}
			if bomb == nil {
				g.Surge()
				return
			}
			g.Logf("Explosion: %d damage to assign.", bomb.Threat)
			assignDamage(g, "Explosion", bomb.Threat)
		}},
	})
	add(&e.CardDef{
		Code: "01112", Name: "False Alarm", Type: e.TypeTreachery, Boost: 1,
		Text: "You are confused. If you already are, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if g.S.Hero.Confused {
				g.Surge()
				return
			}
			g.Confuse(g.S.Hero)
		}},
	})

	// --- Spider-Man obligation and nemesis set ---------------------------------------
	add(&e.CardDef{
		Code: "01165", Name: "Eviction Notice", Type: e.TypeObligation, Boost: 2,
		Text: "Choose: exhaust Peter Parker (flipping to alter-ego if needed) to remove this card from the game, or discard a random card from your hand and surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			s := g.S
			var opts []e.Option
			if !s.Hero.Exhausted {
				text := "Exhaust Peter Parker: remove Eviction Notice from the game"
				if g.IsHero() {
					text = "Flip to Peter Parker and exhaust him: remove Eviction Notice from the game"
				}
				opts = append(opts, e.NewAction("effect:Eviction Notice>exhaust", text, func() {
					if g.IsHero() {
						from := s.Hero.Name()
						s.Hero.Flipped = true
						g.Logf("%s changes form to %s.", from, s.Hero.Name())
					}
					s.Hero.Exhausted = true
					g.RemoveFromGame(c)
				}))
			}
			opts = append(opts, e.NewAction("effect:Eviction Notice>discard", "Discard a random card from your hand; Eviction Notice surges", func() {
				if x := g.RandomHandCard(); x != nil {
					g.DiscardFromHand(x)
				}
				s.EncDiscard = append(s.EncDiscard, c)
				g.Surge()
			}))
			g.Do(func() {
				g.Ask(&e.Decision{Kind: "choice", Prompt: "Eviction Notice (obligation): choose how to resolve it.", Options: opts})
			})
		}},
	})
	add(&e.CardDef{
		Code: "01166", Name: "Highway Robbery", Type: e.TypeSideScheme, Boost: 3, Acceleration: 1, StartingThreat: 3,
		Text: "Acceleration. When revealed, a random card from your hand is placed here face-down; it returns to your hand when this scheme is defeated.",
		Script: &e.Script{
			OnReveal: func(g *e.Game, c *e.Card) {
				if x := g.RandomHandCard(); x != nil {
					g.S.Hand = without(g.S.Hand, x)
					c.Attached = append(c.Attached, x)
					g.Logf("Highway Robbery takes %s from your hand.", x.Def.Name)
				}
			},
			OnDefeated: func(g *e.Game, c *e.Card) {
				for _, x := range c.Attached {
					g.S.Hand = append(g.S.Hand, x)
					g.Logf("%s returns to your hand.", x.Def.Name)
				}
				c.Attached = nil
			},
		},
	})
	add(&e.CardDef{Code: "01167", Name: "Vulture", Type: e.TypeMinion, Unique: true, HP: 4, ATK: 3, SCH: 1, Boost: 2, Quickstrike: true,
		Text: "Quickstrike (attacks you as soon as he engages you in hero form)."})
	add(&e.CardDef{
		Code: "01168", Name: "Sweeping Swoop", Type: e.TypeTreachery,
		Text: "Your identity is stunned. If Vulture is in play, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			g.Stun(g.S.Hero)
			for _, m := range g.S.Minions {
				if m.Def.Code == "01167" {
					g.Surge()
					return
				}
			}
		}},
	})
	add(&e.CardDef{
		Code: "01169", Name: "The Vulture's Plans", Type: e.TypeTreachery, Boost: 2,
		Text: "Discard a random card from your hand; place 1 threat on the main scheme for each different resource type on it.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			x := g.RandomHandCard()
			if x == nil {
				return
			}
			g.DiscardFromHand(x)
			kinds := map[e.Resource]bool{}
			for _, r := range x.Def.Resources {
				kinds[r] = true
			}
			g.PlaceThreat(g.S.MainScheme, len(kinds))
		}},
	})
}

// controlledUpgradesSupports lists upgrades and supports you control,
// including your upgrades attached to enemies.
func controlledUpgradesSupports(g *e.Game) []*e.Card {
	var out []*e.Card
	isUS := func(c *e.Card) bool { return c.Def.Type == e.TypeUpgrade || c.Def.Type == e.TypeSupport }
	for _, c := range g.S.Play {
		if isUS(c) {
			out = append(out, c)
		}
	}
	for _, host := range append(g.AllEnemies(), g.S.Hero) {
		for _, a := range host.Attached {
			if isUS(a) {
				out = append(out, a)
			}
		}
	}
	return out
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
		opts := []e.Option{e.NewAction("effect:"+source+">"+g.S.Hero.Name(), "Deal 1 to "+g.S.Hero.Name(), func() {
			g.DamageHero(1)
			assignDamage(g, source, n-1)
		})}
		for _, a := range allies {
			ally := a
			opts = append(opts, e.NewAction("effect:"+source+">"+g.Label(ally), fmt.Sprintf("Deal 1 to %s (%d HP left)", g.Label(ally), ally.RemainingHP()), func() {
				g.DamageAlly(ally, 1)
				assignDamage(g, source, n-1)
			}))
		}
		g.Ask(&e.Decision{Kind: "choice", Prompt: fmt.Sprintf("%s: assign 1 damage (%d left to assign).", source, n), Options: opts})
	})
}

func contains(zone []*e.Card, c *e.Card) bool {
	for _, x := range zone {
		if x == c {
			return true
		}
	}
	return false
}

func without(zone []*e.Card, c *e.Card) []*e.Card {
	out := make([]*e.Card, 0, len(zone))
	for _, x := range zone {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

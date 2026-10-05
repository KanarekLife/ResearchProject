package cards

import (
	"fmt"

	e "mcbench/engine"
)

// Remaining cards of the Spider-Man starter deck (Justice + basic).

func init() {
	add(&e.CardDef{
		Code: "01058", Name: "Daredevil", Type: e.TypeAlly, Unique: true, Cost: 4, Resources: rs(P),
		HP: 3, ATK: 2, AtkCons: 1, THW: 2, ThwCons: 1,
		Text: "After Daredevil thwarts, deal 1 damage to an enemy.",
		Script: &e.Script{Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
			e.TrigThwarted: func(g *e.Game, c *e.Card, ev *e.Event) {
				if ev.Source == c {
					g.ChooseEnemy("Daredevil", "Deal 1 damage to", func(t *e.Card) { g.DamageEnemy(t, 1) })
				}
			},
		}},
	})
	add(&e.CardDef{
		Code: "01059", Name: "Jessica Jones", Type: e.TypeAlly, Unique: true, Cost: 3, Resources: rs(E),
		HP: 3, ATK: 2, AtkCons: 1, THW: 1, ThwCons: 1,
		Text: "Gets +1 THW for each side scheme in play.",
		Script: &e.Script{StatBonus: func(g *e.Game, c *e.Card, stat string) int {
			if stat == "thw" {
				return len(g.S.SideSchemes)
			}
			return 0
		}},
	})
	add(&e.CardDef{
		Code: "01061", Name: "Great Responsibility", Type: e.TypeEvent, Cost: 0, Resources: rs(M),
		Text: "Hero Interrupt: when threat would be placed on a scheme during the villain phase, take that much damage instead.",
		Script: &e.Script{
			PlayWindow: e.TrigThreatWouldBePlaced,
			PlayIf:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.IsHero() && ev.Amount > 0 },
			OnPlayEvent: func(g *e.Game, c *e.Card, ev *e.Event) {
				n := ev.Amount
				ev.Amount = 0
				g.Logf("Great Responsibility: %d threat is taken as damage instead.", n)
				g.DamageHero(n)
			},
		},
	})
	add(&e.CardDef{
		Code: "01062", Name: "The Power of Justice", Type: e.TypeResource, Resources: rs(W), DoubleFor: "Justice",
		Text: "Generates 2 wild resources when paying for a Justice card.",
	})
	add(&e.CardDef{
		Code: "01063", Name: "Interrogation Room", Type: e.TypeSupport, Cost: 1, Resources: rs(E), MaxInPlay: 1,
		Text: "Max 1. After a minion is defeated, exhaust Interrogation Room to remove 1 threat from a scheme (used automatically).",
		Script: &e.Script{Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
			e.TrigMinionDefeated: func(g *e.Game, c *e.Card, ev *e.Event) {
				if !c.Exhausted && len(g.ThwartableSchemes()) > 0 {
					c.Exhausted = true
					g.ChooseScheme("Interrogation Room", 1)
				}
			},
		}},
	})
	add(&e.CardDef{
		Code: "01064", Name: "Surveillance Team", Type: e.TypeSupport, Cost: 2, Resources: rs(M), Uses: 3,
		Text: "Uses 3 counters. Action: exhaust and remove a counter to remove 1 threat from a scheme; discard when empty.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if c.Exhausted || c.Counters == 0 {
				return nil
			}
			var out []e.Option
			for _, sc := range g.ThwartableSchemes() {
				scheme := sc
				out = append(out, e.NewAction("use:Surveillance Team>"+g.Label(scheme), "Use Surveillance Team: remove 1 threat from "+g.Label(scheme)+" (exhaust, remove a counter)", func() {
					c.Exhausted = true
					c.Counters--
					g.RemoveThreat(scheme, 1)
					if c.Counters == 0 {
						g.Detach(c)
					}
				}))
			}
			return out
		}},
	})
	add(&e.CardDef{
		Code: "01065", Name: "Heroic Intuition", Type: e.TypeUpgrade, Cost: 2, Resources: rs(E), MaxInPlay: 1, HeroTHW: 1,
		Text: "Max 1. Your hero gets +1 THW.",
	})
	add(&e.CardDef{
		Code: "01084", Name: "Nick Fury", Type: e.TypeAlly, Unique: true, Cost: 4, Resources: rs(M),
		HP: 3, ATK: 2, AtkCons: 1, THW: 2, ThwCons: 1,
		Text: "When he enters play, choose: remove 2 threat from a scheme, draw 3 cards, or deal 4 damage to an enemy. Discarded at the end of the round.",
		Script: &e.Script{
			OnPlay: func(g *e.Game, c *e.Card, _ *e.Card) {
				g.Do(func() {
					opts := []e.Option{
						e.NewAction("effect:Nick Fury>draw", "Draw 3 cards", func() { g.Draw(3) }),
						e.NewAction("effect:Nick Fury>damage", "Deal 4 damage to an enemy", func() {
							g.ChooseEnemy("Nick Fury", "Deal 4 damage to", func(t *e.Card) { g.DamageEnemy(t, 4) })
						}),
					}
					if len(g.ThwartableSchemes()) > 0 {
						opts = append(opts, e.NewAction("effect:Nick Fury>thwart", "Remove 2 threat from a scheme", func() { g.ChooseScheme("Nick Fury", 2) }))
					}
					g.Ask(&e.Decision{Kind: "choice", Prompt: "Nick Fury enters play: choose one effect.", Options: opts})
				})
			},
			Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
				e.TrigRoundEnd: func(g *e.Game, c *e.Card, ev *e.Event) {
					g.Logf("Nick Fury leaves at the end of the round.")
					g.Detach(c)
				},
			},
		},
	})
	add(&e.CardDef{
		Code: "01085", Name: "Emergency", Type: e.TypeEvent, Cost: 0, Resources: rs(E),
		Text: "Interrupt: when the villain schemes, reduce the threat it places by 1.",
		Script: &e.Script{
			PlayWindow: e.TrigVillainSchemes,
			PlayIf:     func(g *e.Game, c *e.Card, ev *e.Event) bool { return ev.Amount > 0 },
			OnPlayEvent: func(g *e.Game, c *e.Card, ev *e.Event) {
				ev.Amount--
				g.Logf("Emergency reduces the scheme by 1.")
			},
		},
	})
	add(&e.CardDef{
		Code: "01091", Name: "Avengers Mansion", Type: e.TypeSupport, Cost: 4, Resources: rs(M), MaxInPlay: 1,
		Text: "Max 1. Action: exhaust to draw 1 card.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if c.Exhausted {
				return nil
			}
			return []e.Option{e.NewAction("use:Avengers Mansion", "Use Avengers Mansion: draw 1 card (exhaust)", func() {
				c.Exhausted = true
				g.Draw(1)
			})}
		}},
	})
	add(&e.CardDef{
		Code: "01092", Name: "Helicarrier", Type: e.TypeSupport, Cost: 3, Resources: rs(P), MaxInPlay: 1,
		Text: "Max 1. Action: exhaust to reduce the cost of the next card you play this phase by 1.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if c.Exhausted {
				return nil
			}
			return []e.Option{e.NewAction("use:Helicarrier", "Use Helicarrier: the next card you play this phase costs 1 less (exhaust)", func() {
				c.Exhausted = true
				g.S.CostReduction++
				g.Logf("Helicarrier: the next card costs %d less.", g.S.CostReduction)
			})}
		}},
	})
	add(&e.CardDef{
		Code: "01093", Name: "Tenacity", Type: e.TypeUpgrade, Cost: 2, Resources: rs(E),
		Text: "Hero Action: spend a physical resource and discard Tenacity to ready your hero.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if !g.IsHero() || !g.S.Hero.Exhausted {
				return nil
			}
			return g.NewActionPaid(c, "use:Tenacity", "Use Tenacity: ready your hero (discard Tenacity)", 1, rs(P), func() {
				g.Detach(c)
				g.S.Hero.Exhausted = false
				g.Logf("Tenacity readies %s.", g.S.Hero.Name())
			})
		}},
	})
}

// aspectOf assigns the player-card class of core-set cards by code range.
func aspectOf(code string) string {
	var n int
	fmt.Sscanf(code, "%d", &n)
	switch {
	case n >= 1001 && n <= 1047:
		return "Hero"
	case n >= 1050 && n <= 1057:
		return "Aggression"
	case n >= 1058 && n <= 1065:
		return "Justice"
	case n >= 1066 && n <= 1074:
		return "Leadership"
	case n >= 1075 && n <= 1082:
		return "Protection"
	case n >= 1083 && n <= 1093:
		return "Basic"
	}
	return ""
}

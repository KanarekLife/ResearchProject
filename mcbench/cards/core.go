package cards

import (
	"slices"

	e "mcbench/engine"
)

// Core Set: Spider-Man, a few aspect/basic cards, and the Rhino scenario.
// Stats follow the printed cards for one player; Text is our own paraphrase.

var (
	P = e.Physical
	M = e.Mental
	E = e.Energy
	W = e.Wild
)

func init() {
	// --- Spider-Man -------------------------------------------------------------
	peter := &e.CardDef{
		Code: "01001b", Name: "Peter Parker", Type: e.TypeAlterEgo, Unique: true,
		HP: 10, REC: 3, HandSize: 6,
		Text: "Scientist - Resource: generate a mental resource (once per round).",
		Script: &e.Script{ResourceAbility: &e.ResourceAbility{
			Gives:  M,
			Usable: func(g *e.Game, c *e.Card) bool { return !g.S.OncePerRound["peter-parker"] },
			Use: func(g *e.Game, c *e.Card) {
				g.S.OncePerRound["peter-parker"] = true
				g.Logf("Peter Parker generates a mental resource.")
			},
		}},
	}
	add(&e.CardDef{
		Code: "01001a", Name: "Spider-Man", Type: e.TypeHero, Unique: true,
		HP: 10, ATK: 2, THW: 1, DEF: 3, HandSize: 5, Back: peter,
		Text: "Spider-Sense - when the villain attacks you, draw 1 card (applied automatically).",
		Script: &e.Script{Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
			e.TrigVillainAttacks: func(g *e.Game, c *e.Card, ev *e.Event) {
				g.Logf("Spider-Sense!")
				g.Draw(1)
			},
		}},
	})
	add(&e.CardDef{
		Code: "01002", Name: "Black Cat", Type: e.TypeAlly, Unique: true, Cost: 2, Resources: rs(E),
		HP: 2, ATK: 1, THW: 1, ThwCons: 1,
		Text: "When played: discard the top 2 cards of your deck; put each one with a printed mental resource into your hand.",
		Script: &e.Script{OnPlay: func(g *e.Game, c *e.Card, _ *e.Card) {
			for i := 0; i < 2 && len(g.S.Deck) > 0; i++ {
				top := g.S.Deck[0]
				g.S.Deck = g.S.Deck[1:]
				if slices.Contains(top.Def.Resources, M) {
					g.S.Hand = append(g.S.Hand, top)
					g.Logf("Black Cat finds %s (added to hand).", top.Def.Name)
				} else {
					g.S.Discard = append(g.S.Discard, top)
					g.Logf("Black Cat discards %s.", top.Def.Name)
				}
			}
		}},
	})
	add(&e.CardDef{
		Code: "01003", Name: "Backflip", Type: e.TypeEvent, Cost: 0, Resources: rs(P),
		Text: "Interrupt: when you would take damage from an attack, prevent all of it.",
		Script: &e.Script{
			PlayWindow:  e.TrigWouldTakeAttackDamage,
			OnPlayEvent: func(g *e.Game, c *e.Card, ev *e.Event) { ev.Amount = 0; g.Logf("Backflip prevents the damage.") },
		},
	})
	add(&e.CardDef{
		Code: "01004", Name: "Enhanced Spider-Sense", Type: e.TypeEvent, Cost: 1, Resources: rs(M),
		Text: "Hero Interrupt: when a treachery is revealed, cancel its When Revealed effects.",
		Script: &e.Script{
			PlayWindow:  e.TrigTreacheryRevealed,
			PlayIf:      func(g *e.Game, c *e.Card, ev *e.Event) bool { return g.IsHero() },
			OnPlayEvent: func(g *e.Game, c *e.Card, ev *e.Event) { ev.Cancelled = true },
		},
	})
	add(heroAttackEvent("01005", "Swinging Web Kick", 3, M, 8))
	add(&e.CardDef{
		Code: "01006", Name: "Aunt May", Type: e.TypeSupport, Unique: true, Cost: 1, Resources: rs(E),
		Text: "Alter-Ego Action: exhaust Aunt May to heal 4 damage from Peter Parker.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if g.IsHero() || c.Exhausted || g.S.Hero.Damage == 0 {
				return nil
			}
			return []e.Option{e.NewAction("use:Aunt May", "Use Aunt May: heal 4 damage from Peter Parker (exhaust Aunt May)", func() {
				c.Exhausted = true
				g.HealHero(4)
			})}
		}},
	})
	add(&e.CardDef{
		Code: "01007", Name: "Spider-Tracer", Type: e.TypeUpgrade, Cost: 1, Resources: rs(E),
		Text: "Attach to a minion. When that minion is defeated, remove 3 threat from a scheme.",
		Script: &e.Script{
			Targets: func(g *e.Game, c *e.Card) []*e.Card { return g.S.Minions },
			Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
				e.TrigMinionDefeated: func(g *e.Game, c *e.Card, ev *e.Event) {
					if g.HostOf(c) == ev.Target {
						g.ChooseScheme("Spider-Tracer", 3)
					}
				},
			},
		},
	})
	add(&e.CardDef{
		Code: "01008", Name: "Web-Shooter", Type: e.TypeUpgrade, Cost: 1, Resources: rs(P), Uses: 3,
		Text: "Uses 3 web counters. Hero Resource: exhaust and remove a counter to generate a wild resource; discard when empty.",
		Script: &e.Script{ResourceAbility: &e.ResourceAbility{
			Gives:  W,
			Usable: func(g *e.Game, c *e.Card) bool { return g.IsHero() && !c.Exhausted && c.Counters > 0 },
			Use: func(g *e.Game, c *e.Card) {
				c.Exhausted = true
				c.Counters--
				g.Logf("Web-Shooter generates a wild resource (%d counters left).", c.Counters)
				if c.Counters == 0 {
					g.Detach(c)
				}
			},
		}},
	})
	add(&e.CardDef{
		Code: "01009", Name: "Webbed Up", Type: e.TypeUpgrade, Cost: 4, Resources: rs(P),
		Text: "Hero form only. Attach to an enemy (max 1 per enemy). When that enemy would attack, discard Webbed Up instead and stun the enemy.",
		Script: &e.Script{
			CanPlay: func(g *e.Game, c *e.Card) bool { return g.IsHero() },
			Targets: func(g *e.Game, c *e.Card) []*e.Card {
				var out []*e.Card
				for _, en := range g.AllEnemies() {
					if !slices.ContainsFunc(en.Attached, func(a *e.Card) bool { return a.Def.Name == "Webbed Up" }) {
						out = append(out, en)
					}
				}
				return out
			},
			Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
				e.TrigEnemyWouldAttack: func(g *e.Game, c *e.Card, ev *e.Event) {
					if host := g.HostOf(c); host == ev.Source {
						ev.Cancelled = true
						g.Logf("Webbed Up stops %s's attack.", host.Name())
						g.Detach(c)
						g.Stun(host)
					}
				},
			},
		},
	})

	// --- Aspect and basic cards -------------------------------------------------
	add(heroAttackEvent("01087", "Haymaker", 2, E, 3))
	add(&e.CardDef{
		Code: "01060", Name: "For Justice!", Type: e.TypeEvent, Cost: 2, Resources: rs(E),
		Text: "Hero Action (thwart): remove 3 threat from a scheme (4 if you paid with a mental resource).",
		Script: &e.Script{
			CanPlay: func(g *e.Game, c *e.Card) bool { return g.IsHero() },
			Targets: func(g *e.Game, c *e.Card) []*e.Card { return g.ThwartableSchemes() },
			OnPlay: func(g *e.Game, c *e.Card, t *e.Card) {
				n := 3
				if slices.Contains(g.LastPaid, M) {
					n = 4
				}
				g.Thwart(g.S.Hero, t, n, 0)
			},
		},
	})
	add(&e.CardDef{
		Code: "01086", Name: "First Aid", Type: e.TypeEvent, Cost: 1, Resources: rs(M),
		Text: "Action: heal 2 damage from your identity.",
		Script: &e.Script{
			CanPlay: func(g *e.Game, c *e.Card) bool { return g.S.Hero.Damage > 0 },
			OnPlay:  func(g *e.Game, c *e.Card, _ *e.Card) { g.HealHero(2) },
		},
	})
	add(&e.CardDef{
		Code: "01083", Name: "Mockingbird", Type: e.TypeAlly, Unique: true, Cost: 3, Resources: rs(P),
		HP: 3, ATK: 1, AtkCons: 1, THW: 1, ThwCons: 1,
		Text: "When she enters play, stun an enemy.",
		Script: &e.Script{OnPlay: func(g *e.Game, c *e.Card, _ *e.Card) {
			g.ChooseEnemy("Mockingbird", "Stun", g.Stun)
		}},
	})
	add(&e.CardDef{Code: "01088", Name: "Energy", Type: e.TypeResource, Resources: rs(E, E)})
	add(&e.CardDef{Code: "01089", Name: "Genius", Type: e.TypeResource, Resources: rs(M, M)})
	add(&e.CardDef{Code: "01090", Name: "Strength", Type: e.TypeResource, Resources: rs(P, P)})

	// --- Rhino scenario -------------------------------------------------------
	add(&e.CardDef{Code: "01094", Name: "Rhino", Type: e.TypeVillain, Unique: true, HP: 14, ATK: 2, SCH: 1})
	add(&e.CardDef{
		Code: "01095", Name: "Rhino", Type: e.TypeVillain, Unique: true, HP: 15, ATK: 3, SCH: 1,
		Text: "When revealed: find Breakin' & Takin' in the encounter deck or discard and reveal it; shuffle the encounter deck.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			s := g.S
			for _, zone := range []*[]*e.Card{&s.EncDeck, &s.EncDiscard} {
				for _, x := range *zone {
					if x.Def.Code == "01107" {
						*zone = slices.DeleteFunc(*zone, func(y *e.Card) bool { return y == x })
						g.Shuffle(s.EncDeck)
						g.Do(func() { g.Reveal(x) })
						return
					}
				}
			}
			g.Shuffle(s.EncDeck)
		}},
	})
	add(&e.CardDef{
		Code: "01096", Name: "Rhino", Type: e.TypeVillain, Unique: true, HP: 16, ATK: 4, SCH: 1, Toughness: true,
		Text:   "Toughness. When revealed: stun your hero.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) { g.Stun(g.S.Hero) }},
	})
	add(&e.CardDef{
		Code: "01097", Name: "The Break-In!", Type: e.TypeMainScheme, TargetThreat: 7, Escalation: 1,
		Text: "If threat reaches the target, you lose.",
	})
	add(&e.CardDef{
		Code: "01098", Name: "Armored Rhino Suit", Type: e.TypeAttachment,
		Text: "Damage that would be dealt to Rhino is placed here instead; discarded once it holds 5 or more.",
		Script: &e.Script{Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
			e.TrigEnemyWouldTakeDamage: func(g *e.Game, c *e.Card, ev *e.Event) {
				if g.HostOf(c) != ev.Target || ev.Amount <= 0 {
					return
				}
				c.Damage += ev.Amount
				g.Logf("Armored Rhino Suit absorbs %d damage (%d on it).", ev.Amount, c.Damage)
				ev.Amount = 0
				if c.Damage >= 5 {
					g.Logf("Armored Rhino Suit is discarded.")
					c.Damage = 0
					g.Detach(c)
				}
			},
		}},
	})
	add(&e.CardDef{
		Code: "01099", Name: "Charge", Type: e.TypeAttachment, Boost: 2, AttachATK: 3, AttachOverkill: true,
		Text: "Rhino gets +3 ATK and his attacks gain overkill. Discarded after Rhino's next attack.",
		Script: &e.Script{Forced: map[e.Trigger]func(*e.Game, *e.Card, *e.Event){
			e.TrigAttackEnded: func(g *e.Game, c *e.Card, ev *e.Event) {
				if g.HostOf(c) == ev.Source {
					g.Logf("Charge is discarded.")
					g.Detach(c)
				}
			},
		}},
	})
	add(&e.CardDef{
		Code: "01100", Name: "Enhanced Ivory Horn", Type: e.TypeAttachment, Boost: 2, AttachATK: 1,
		Text: "Rhino gets +1 ATK. Hero Action: spend 3 physical resources to discard this card.",
		Script: &e.Script{Actions: func(g *e.Game, c *e.Card) []e.Option {
			if !g.IsHero() {
				return nil
			}
			return g.NewActionPaid(c, "use:Enhanced Ivory Horn", "Discard Enhanced Ivory Horn by spending 3 physical resources", 3, rs(P, P, P), func() {
				g.Logf("Enhanced Ivory Horn is discarded.")
				g.Detach(c)
			})
		}},
	})
	add(&e.CardDef{Code: "01101", Name: "Hydra Mercenary", Type: e.TypeMinion, HP: 3, ATK: 1, SCH: 0, Boost: 1, Guard: true,
		Text: "Guard (while engaged with you, you cannot attack the villain)."})
	add(&e.CardDef{Code: "01102", Name: "Sandman", Type: e.TypeMinion, Unique: true, HP: 4, ATK: 3, SCH: 2, Boost: 2, Toughness: true,
		Text: "Toughness."})
	add(&e.CardDef{
		Code: "01103", Name: "Shocker", Type: e.TypeMinion, Unique: true, HP: 3, ATK: 2, SCH: 1, Boost: 2,
		Text:   "When revealed: deal 1 damage to your hero.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) { g.DamageHero(1) }},
	})
	add(&e.CardDef{
		Code: "01104", Name: "Hard to Keep Down", Type: e.TypeTreachery,
		Text: "Rhino heals 4 damage. If nothing was healed, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if g.HealVillain(4) == 0 {
				g.Surge()
			}
		}},
	})
	add(&e.CardDef{
		Code: "01105", Name: "I'm Tough", Type: e.TypeTreachery,
		Text: "Give Rhino a tough status. If he already has one, surge.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if g.S.Villain.Tough {
				g.Surge()
			} else {
				g.GiveTough(g.S.Villain)
			}
		}},
	})
	add(&e.CardDef{
		Code: "01106", Name: "Stampede", Type: e.TypeTreachery, Boost: 1,
		Text: "Alter-ego: surge. Hero: Rhino attacks you; a character damaged by this attack is stunned.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			if !g.IsHero() {
				g.Surge()
				return
			}
			g.EnemyAttack(g.S.Villain, true, g.Stun)
		}},
	})
	add(&e.CardDef{Code: "01107", Name: "Breakin' & Takin'", Type: e.TypeSideScheme, Boost: 2, StartingThreat: 3, Hazard: 1,
		Text: "Hazard: deal 1 extra encounter card each villain phase."})
	add(&e.CardDef{Code: "01108", Name: "Crowd Control", Type: e.TypeSideScheme, Boost: 2, StartingThreat: 2, Crisis: true,
		Text: "Crisis: threat cannot be removed from the main scheme."})
	add(&e.CardDef{Code: "01109", Name: "Bomb Scare", Type: e.TypeSideScheme, Boost: 2, StartingThreat: 3, Acceleration: 1,
		Text: "Acceleration: +1 threat on the main scheme each villain phase."})
	add(&e.CardDef{
		Code: "01110", Name: "Hydra Bomber", Type: e.TypeMinion, HP: 2, ATK: 1, SCH: 1, Boost: 1,
		Text: "When revealed: you either take 2 damage or place 1 threat on the main scheme.",
		Script: &e.Script{OnReveal: func(g *e.Game, c *e.Card) {
			g.Do(func() {
				g.Ask(&e.Decision{Kind: "choice", Prompt: "Hydra Bomber: take 2 damage, or place 1 threat on the main scheme?", Options: []e.Option{
					e.NewAction("effect:Hydra Bomber>damage", "Take 2 damage", func() { g.DamageHero(2) }),
					e.NewAction("effect:Hydra Bomber>threat", "Place 1 threat on the main scheme", func() { g.PlaceThreat(g.S.MainScheme, 1) }),
				}})
			})
		}},
	})
}

// heroAttackEvent is a "Hero Action (attack): deal N damage to an enemy" event.
func heroAttackEvent(code, name string, cost int, res e.Resource, dmg int) *e.CardDef {
	return &e.CardDef{
		Code: code, Name: name, Type: e.TypeEvent, Cost: cost, Resources: rs(res),
		Text: "Hero Action (attack): deal " + itoa(dmg) + " damage to an enemy.",
		Script: &e.Script{
			CanPlay: func(g *e.Game, c *e.Card) bool { return g.IsHero() },
			Targets: func(g *e.Game, c *e.Card) []*e.Card { return g.Enemies() },
			OnPlay:  func(g *e.Game, c *e.Card, t *e.Card) { g.Attack(g.S.Hero, t, dmg, 0) },
		},
	}
}

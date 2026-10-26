package cards

import (
	"testing"

	"mcbench/constants"
	e "mcbench/game/engine"
)

// Rule checks for the mechanics the Iron Man, Captain Marvel, Klaw and
// Ultron cards need. Positions are built by hand on top of newGameWith.

func TestMainSchemeAdvancesBeforeTheFinalStageLoses(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.MainScheme = mk("Underground Distribution")
		s.MainScheme.Threat = 5
		next, _ := Get("Secret Rendezvous")
		s.MainStages = []*e.CardDef{next}
		s.EncDeck = append([]*e.Card{mk("Advance"), mk("Hydra Mercenary")}, s.EncDeck...)
	})
	choose(t, g, "end_turn")
	if g.Result().Over || g.S.MainScheme.Name() != "Secret Rendezvous" {
		t.Fatalf("over %v scheme %s; want the scheme to advance, not a loss", g.Result().Over, g.S.MainScheme.Name())
	}
	m := g.Metrics()
	if m[constants.MetricSchemeTotal] != 14 || m[constants.MetricSchemeProgress] != 6+float64(g.S.MainScheme.Threat) {
		t.Fatalf("progress %v/%v; want the completed stage's 6 counted", m[constants.MetricSchemeProgress], m[constants.MetricSchemeTotal])
	}
}

func TestDefeatedDroneGoesToYourDiscard(t *testing.T) {
	var top *e.Card
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		top = mk("Genius")
		top.Drone = droneFace()
		s.Minions = []*e.Card{top}
		s.Deck = []*e.Card{mk("Energy")}
		s.Environment = []*e.Card{mk("Ultron Drones")}
		s.Environment[0].Attached = []*e.Card{mk("Upgraded Drones")}
	})
	if g.MaxHP(top) != 2 || g.EnemyATK(top) != 2 {
		t.Fatalf("drone HP %d ATK %d; want 2 and 2 with Upgraded Drones", g.MaxHP(top), g.EnemyATK(top))
	}
	choose(t, g, "attack:Spider-Man>Drone")
	if len(g.S.Minions) != 0 || len(g.S.Discard) != 1 || g.S.Discard[0] != top || top.Drone != nil {
		t.Fatalf("minions %d discard %v; want the Genius back in your discard pile", len(g.S.Minions), g.S.Discard)
	}
}

func TestRetaliateDamagesTheAttacker(t *testing.T) {
	g := newGame(t, nil, "Whiplash")
	choose(t, g, "attack:Spider-Man>Whiplash")
	if g.S.Hero.Damage != 1 {
		t.Fatalf("hero damage %d; want 1 from retaliate", g.S.Hero.Damage)
	}
}

func TestKlawGetsAnAdditionalBoostCard(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Villain = mk("01113")
		s.EncDeck = []*e.Card{mk("Hydra Mercenary"), mk("Hydra Mercenary"), mk("Hydra Mercenary"), mk("Hydra Mercenary")}
	})
	choose(t, g, "end_turn")
	choose(t, g, "pass") // Spider-Sense
	choose(t, g, "defend:none")
	if len(g.S.EncDeck) != 1 {
		t.Fatalf("encounter deck %d; want 2 boost cards and 1 encounter card dealt", len(g.S.EncDeck))
	}
}

func TestIronManHandSizeCountsTechUpgrades(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Hero = mk("Iron Man")
		s.Play = []*e.Card{mk("Mark V Armor"), mk("Rocket Boots"), mk("Combat Training")}
	})
	if g.HandSize() != 3 || g.MaxHP(g.S.Hero) != 16 {
		t.Fatalf("hand size %d HP %d; want 1+2 Tech upgrades and 9+6+1", g.HandSize(), g.MaxHP(g.S.Hero))
	}
}

func TestLosingImmortalKlawCanDefeatTheStage(t *testing.T) {
	g := newGameWith(t, []string{"Swinging Web Kick", "Genius", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		s.Villain = mk("01114")
		s.Villain.Damage = 20
		immortal := mk(`The "Immortal" Klaw`)
		immortal.Threat = 1
		s.SideSchemes = []*e.Card{immortal}
		next, _ := Get("01115")
		s.VillainStages = []*e.CardDef{next}
	})
	choose(t, g, `thwart:Spider-Man>The "Immortal" Klaw`)
	if g.S.Villain.Def.Code != "01115" {
		t.Fatalf("villain %s; want stage II defeated once its +10 hit points are gone", g.S.Villain.Def.Code)
	}
}

func TestOptionalResponsesAsk(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Play = []*e.Card{mk("Hawkeye")}
		s.Play[0].Counters = 4
		s.Villain.Stunned = true
		s.EncDeck = []*e.Card{mk("Hydra Mercenary"), mk(filler)}
	})
	choose(t, g, "end_turn")
	if !hasKey(g, "use:Hawkeye") || !hasKey(g, constants.KeyPass) {
		t.Fatalf("want Hawkeye's response offered with a pass, got %v", keys(g))
	}
	choose(t, g, "use:Hawkeye")
	if m := g.S.Minions[0]; m.Damage != 2 || g.S.Play[0].Counters != 3 {
		t.Fatalf("minion damage %d arrows %d; want 2 and 3", m.Damage, g.S.Play[0].Counters)
	}
}

func TestCosmicFlightPreventsDamageFromAnySource(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Hero = mk("Captain Marvel")
		s.Play = []*e.Card{mk("Cosmic Flight")}
	})
	g.HarmHero(4, nil)
	g.Run()
	choose(t, g, "use:Cosmic Flight")
	if g.S.Hero.Damage != 1 || len(g.S.Play) != 0 {
		t.Fatalf("damage %d in play %d; want 1 damage and Cosmic Flight discarded", g.S.Hero.Damage, len(g.S.Play))
	}
}

func TestMakeTheCallPaysTheAllysCost(t *testing.T) {
	// The Power of Leadership doubles for the Leadership ally (Rules Reference FAQ).
	g := newGameWith(t, []string{"Make the Call", "The Power of Leadership"}, func(s *e.State, mk func(string) *e.Card) {
		s.Deck = []*e.Card{mk("Energy")}
		s.Discard = []*e.Card{mk("Maria Hill")}
	})
	choose(t, g, "play:Make the Call>Maria Hill|pay:The Power of Leadership")
	choose(t, g, "pass") // Maria Hill's draw
	if len(g.S.Play) != 1 || g.S.Play[0].Def.Name != "Maria Hill" {
		t.Fatalf("in play %v; want Maria Hill from the discard pile", g.S.Play)
	}
}

func TestRelentlessAssaultOverkill(t *testing.T) {
	g := newGameWith(t, []string{"Relentless Assault", "Strength"}, func(s *e.State, mk func(string) *e.Card) {
		s.Minions = []*e.Card{mk("Hydra Mercenary")}
	})
	choose(t, g, "play:Relentless Assault>Hydra Mercenary|pay:Strength")
	if len(g.S.Minions) != 0 || g.S.Villain.Damage != 2 {
		t.Fatalf("minions %d villain damage %d; want the 2 excess damage on Rhino", len(g.S.Minions), g.S.Villain.Damage)
	}
}

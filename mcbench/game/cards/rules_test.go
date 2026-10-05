package cards

import (
	"slices"
	"strings"
	"testing"

	e "mcbench/game/engine"
)

// newGame builds a minimal hero-form Spider-Man vs Rhino game at the start
// of the player's turn.
func newGame(t *testing.T, hand []string, minions ...string) *e.Game {
	t.Helper()
	id := 0
	mk := func(ref string) *e.Card {
		d, err := Get(ref)
		if err != nil {
			t.Fatal(err)
		}
		id++
		return &e.Card{ID: id, Def: d, Counters: d.Uses}
	}
	s := &e.State{Hero: mk("Spider-Man"), Villain: mk("01094"), MainScheme: mk("The Break-In!")}
	for _, h := range hand {
		s.Hand = append(s.Hand, mk(h))
	}
	for _, m := range minions {
		s.Minions = append(s.Minions, mk(m))
	}
	g := e.New(s, 1)
	g.BeginPlayerTurn()
	return g
}

func keys(g *e.Game) []string {
	var out []string
	for _, o := range g.Pending().Options {
		out = append(out, o.FullKey())
	}
	return out
}

func choose(t *testing.T, g *e.Game, key string) {
	t.Helper()
	for _, o := range g.Pending().Options {
		if o.FullKey() == key {
			if err := g.Choose(o.ID); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("no option %q in %v", key, keys(g))
}

func TestPaymentsAreMinimal(t *testing.T) {
	g := newGame(t, []string{"Swinging Web Kick", "Genius", "Energy", "Backflip"})
	var pays []string
	for _, k := range keys(g) {
		if strings.HasPrefix(k, "play:Swinging Web Kick>Rhino|pay:") {
			pays = append(pays, strings.TrimPrefix(k, "play:Swinging Web Kick>Rhino|pay:"))
		}
	}
	want := []string{"Backflip+Genius", "Backflip+Energy", "Energy+Genius"}
	slices.Sort(pays)
	slices.Sort(want)
	if !slices.Equal(pays, want) {
		t.Fatalf("payments %v, want %v", pays, want)
	}
}

func TestGuardBlocksVillain(t *testing.T) {
	g := newGame(t, nil, "Hydra Mercenary")
	for _, k := range keys(g) {
		if strings.Contains(k, ">Rhino") {
			t.Fatalf("option %q attacks Rhino through Guard", k)
		}
	}
}

func TestToughPreventsDamage(t *testing.T) {
	g := newGame(t, nil)
	g.GiveTough(g.S.Villain)
	choose(t, g, "attack:Spider-Man>Rhino")
	if g.S.Villain.Damage != 0 || g.S.Villain.Tough {
		t.Fatalf("damage %d tough %v; want 0 damage and tough discarded", g.S.Villain.Damage, g.S.Villain.Tough)
	}
}

func TestStunReplacesAttack(t *testing.T) {
	g := newGame(t, nil)
	g.Stun(g.S.Hero)
	choose(t, g, "attack:Spider-Man>Rhino")
	if g.S.Villain.Damage != 0 || g.S.Hero.Stunned || !g.S.Hero.Exhausted {
		t.Fatalf("stun not resolved correctly: dmg %d stunned %v exhausted %v", g.S.Villain.Damage, g.S.Hero.Stunned, g.S.Hero.Exhausted)
	}
}

func TestCrisisBlocksMainSchemeThwart(t *testing.T) {
	g := newGame(t, nil)
	g.S.MainScheme.Threat = 3
	cc, _ := Get("Crowd Control")
	g.S.SideSchemes = append(g.S.SideSchemes, &e.Card{ID: 99, Def: cc, Threat: 2})
	for _, o := range g.TurnOptions() {
		if o.Key == "thwart:Spider-Man>The Break-In!" {
			t.Fatal("main scheme thwartable during crisis")
		}
	}
}

func TestMainSchemeThresholdLoses(t *testing.T) {
	g := newGame(t, nil)
	g.S.MainScheme.Threat = 6
	choose(t, g, "end_turn")
	if r := g.Result(); !r.Over || r.Won {
		t.Fatalf("result %+v, want loss to threat", r)
	}
}

func TestValidateRejectsUnknownPredicates(t *testing.T) {
	bad := &CardDoc{Name: "Bad", Code: "x", Abilities: []Ability{{
		Effects: []Effect{{Verb: "draw", Options: []Choice{{When: []string{"no_such_predicate"}}}}},
	}}}
	if err := validate(bad); err == nil || !strings.Contains(err.Error(), "no_such_predicate") {
		t.Fatalf("validate = %v, want an unknown-predicate error", err)
	}
	ok := &CardDoc{Name: "Ok", Code: "y", Abilities: []Ability{{When: []string{"hero", "paid:mental"}}}}
	if err := validate(ok); err != nil {
		t.Fatal(err)
	}
}

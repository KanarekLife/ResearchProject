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
	return newGameWith(t, hand, func(s *e.State, mk func(string) *e.Card) {
		for _, m := range minions {
			s.Minions = append(s.Minions, mk(m))
		}
	})
}

// filler is the encounter card the test helpers seed the encounter piles with.
const filler = "Advance"

// newGameWith is newGame with prep arranging the state before the turn.
func newGameWith(t *testing.T, hand []string, prep func(s *e.State, mk func(string) *e.Card)) *e.Game {
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
	// Like a real game, the encounter piles are never both empty.
	s.EncDeck = []*e.Card{mk(filler), mk(filler), mk(filler)}
	s.EncDiscard = []*e.Card{mk(filler), mk(filler)}
	for _, h := range hand {
		s.Hand = append(s.Hand, mk(h))
	}
	if prep != nil {
		prep(s, mk)
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

func hasKey(g *e.Game, key string) bool { return slices.Contains(keys(g), key) }

// finishRound answers the villain phase passively (no defense, no
// interrupts) until the next player turn begins.
func finishRound(t *testing.T, g *e.Game) {
	t.Helper()
	round := g.S.Round
	for g.Pending() != nil && g.S.Round == round {
		switch d := g.Pending(); d.Kind {
		case "discard":
			choose(t, g, "discard:done")
		case "defend":
			choose(t, g, "defend:none")
		case "window":
			choose(t, g, "pass")
		default:
			t.Fatalf("unexpected %s decision: %v", d.Kind, keys(g))
		}
	}
}

func TestEndOfTurnMustDiscardDownToHandSize(t *testing.T) {
	g := newGame(t, []string{"Genius", "Genius", "Genius", "Genius", "Genius", "Energy", "Energy"})
	choose(t, g, "end_turn")
	if hasKey(g, "discard:done") {
		t.Fatalf("done offered with 7 cards over hand size 5: %v", keys(g))
	}
	choose(t, g, "discard:Energy")
	choose(t, g, "discard:Energy")
	if !hasKey(g, "discard:done") {
		t.Fatalf("done not offered at hand size: %v", keys(g))
	}
}

func TestForJusticeIsOneThwart(t *testing.T) {
	g := newGameWith(t, []string{"For Justice!", "Genius"}, func(s *e.State, _ func(string) *e.Card) {
		s.MainScheme.Threat = 5
		s.Hero.Confused = true
	})
	choose(t, g, "play:For Justice!>The Break-In!|pay:Genius")
	if g.S.MainScheme.Threat != 5 || g.S.Hero.Confused {
		t.Fatalf("threat %d confused %v; want the whole thwart replaced by the confuse", g.S.MainScheme.Threat, g.S.Hero.Confused)
	}
}

func TestWildPaysForMentalBonus(t *testing.T) {
	g := newGameWith(t, []string{"For Justice!", "The Power of Justice"}, func(s *e.State, _ func(string) *e.Card) {
		s.MainScheme.Threat = 5
	})
	choose(t, g, "play:For Justice!>The Break-In!|pay:The Power of Justice")
	if g.S.MainScheme.Threat != 1 {
		t.Fatalf("threat %d, want 1 (4 removed when paid with a wild resource)", g.S.MainScheme.Threat)
	}
}

func TestDecksResetAsSoonAsEmpty(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Deck = []*e.Card{mk("Genius")}
		s.Discard = []*e.Card{mk("Energy")}
		s.EncDeck = []*e.Card{mk("Advance")}
		s.EncDiscard = []*e.Card{mk("Assault")}
	})
	g.Draw(1)
	s := g.S
	if len(s.Deck) != 1 || len(s.Discard) != 0 || len(s.Dealt) != 1 {
		t.Fatalf("deck %d discard %d dealt %d; want the player deck reset and a card dealt", len(s.Deck), len(s.Discard), len(s.Dealt))
	}
	if len(s.EncDeck) != 1 || s.AccelTokens != 1 {
		t.Fatalf("encounter deck %d tokens %d; want it reset with a token", len(s.EncDeck), s.AccelTokens)
	}
}

func TestMillStopsWhenDeckEmpties(t *testing.T) {
	g := newGameWith(t, []string{"Black Cat", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		s.Deck = []*e.Card{mk("Strength")}
		s.Discard = []*e.Card{mk("Haymaker")}
	})
	choose(t, g, "play:Black Cat|pay:Genius")
	if len(g.S.Deck) != 3 || len(g.S.Discard) != 0 {
		t.Fatalf("deck %d discard %d; want 3 and 0 (no card milled from the new deck)", len(g.S.Deck), len(g.S.Discard))
	}
}

func TestCardsDealtWhileRevealingAreRevealed(t *testing.T) {
	g := newGameWith(t, []string{"Genius", "Genius", "Genius", "Genius", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		s.Villain.Stunned = true
		s.Deck = []*e.Card{mk("Genius")}
		s.Discard = []*e.Card{mk("Energy")}
		s.EncDeck = []*e.Card{mk("Assault"), mk("Advance")}
	})
	choose(t, g, "end_turn")
	finishRound(t, g) // Assault's attack draws (Spider-Sense) and empties the deck
	if len(g.S.Dealt) != 0 || !slices.Contains(g.Log, "Encounter card revealed: Advance (treachery).") {
		t.Fatalf("dealt %d; the card dealt while revealing must be revealed in the same phase", len(g.S.Dealt))
	}
}

func TestLeavingPlayResetsTheCard(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		dd := mk("Daredevil")
		dd.Damage, dd.Exhausted, dd.Stunned, dd.Tough = 2, true, true, true
		s.Play = append(s.Play, dd)
	})
	dd := g.S.Play[0]
	g.Detach(dd)
	if dd.Damage != 0 || dd.Exhausted || dd.Stunned || dd.Tough || !slices.Contains(g.S.Discard, dd) {
		t.Fatalf("discarded ally kept its state: %+v", *dd)
	}
}

func TestFacedownCardsAreOutOfPlay(t *testing.T) {
	g := newGameWith(t, []string{"Avengers Mansion"}, func(s *e.State, mk func(string) *e.Card) {
		s.SideSchemes = append(s.SideSchemes, mk("Highway Robbery"))
	})
	hr := g.S.SideSchemes[0]
	hr.Threat = 3
	hr.Def.Script.OnReveal(g, hr)
	for _, o := range g.TurnOptions() {
		if strings.HasPrefix(o.Key, "use:Avengers Mansion") {
			t.Fatal("the facedown card under Highway Robbery is active")
		}
	}
	g.RemoveThreat(hr, 3)
	if len(g.S.Hand) != 1 || g.S.Hand[0].Facedown {
		t.Fatalf("hand %d; want the card back faceup", len(g.S.Hand))
	}
}

func TestConsequentialDamageAfterResponses(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.MainScheme.Threat = 5
		dd := mk("Daredevil")
		dd.Damage = 2
		s.Play = append(s.Play, dd)
	})
	choose(t, g, "thwart:Daredevil>The Break-In!")
	choose(t, g, "effect:Daredevil>Rhino")
	if g.S.Villain.Damage != 1 || len(g.S.Play) != 0 {
		t.Fatalf("villain damage %d allies %d; want Daredevil's response, then his defeat", g.S.Villain.Damage, len(g.S.Play))
	}
}

func TestAllyLimitDiscardsBeforeEntering(t *testing.T) {
	g := newGameWith(t, []string{"Mockingbird", "Genius", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		s.Play = append(s.Play, mk("Daredevil"), mk("Jessica Jones"), mk("Black Cat"))
	})
	choose(t, g, "play:Mockingbird|pay:Genius+Genius")
	if g.Pending().Kind != "choice" || len(keys(g)) != 4 {
		t.Fatalf("want a choice of the 4 allies to discard, got %v", keys(g))
	}
	choose(t, g, "effect:Mockingbird>Daredevil")
	if !hasKey(g, "effect:Mockingbird>Rhino") {
		t.Fatalf("Mockingbird's stun should follow the ally-limit discard: %v", keys(g))
	}
}

func TestAssignedDamageResolvesTogether(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Villain.Stunned = true
		s.Hero.Tough = true
		bomb := mk("Bomb Scare")
		bomb.Threat = 3
		s.SideSchemes = append(s.SideSchemes, bomb)
		dd := mk("Daredevil")
		dd.Damage = 2
		s.Play = append(s.Play, dd)
		s.EncDeck = []*e.Card{mk("Explosion")}
	})
	choose(t, g, "end_turn")
	choose(t, g, "effect:Explosion>Spider-Man")
	// Daredevil cannot take more than 1, so the last point goes to the hero.
	choose(t, g, "effect:Explosion>Daredevil")
	if g.S.Hero.Damage != 0 || g.S.Hero.Tough || len(g.S.Play) != 0 {
		t.Fatalf("hero damage %d tough %v allies %d; want one tough to prevent all 2", g.S.Hero.Damage, g.S.Hero.Tough, len(g.S.Play))
	}
}

func TestConfusedHeroMayThwartWithoutThreat(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, _ func(string) *e.Card) { s.Hero.Confused = true })
	choose(t, g, "thwart:Spider-Man>The Break-In!")
	if g.S.Hero.Confused || !g.S.Hero.Exhausted {
		t.Fatal("the thwart attempt should exhaust the hero and remove the confuse")
	}
}

func TestGreatResponsibilityOnEncounterThreat(t *testing.T) {
	g := newGameWith(t, []string{"Great Responsibility"}, func(s *e.State, mk func(string) *e.Card) {
		s.Villain.Stunned = true
		s.EncDeck = []*e.Card{mk("Hydra Bomber")}
	})
	choose(t, g, "end_turn")
	choose(t, g, "discard:done")
	choose(t, g, "pass") // the villain phase's own threat
	choose(t, g, "effect:Hydra Bomber>threat")
	choose(t, g, "play:Great Responsibility")
	if g.S.Hero.Damage != 1 || g.S.MainScheme.Threat != 1 {
		t.Fatalf("hero damage %d threat %d; want the encounter threat taken as damage", g.S.Hero.Damage, g.S.MainScheme.Threat)
	}
}

func TestBoostDealtBeforeDefender(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.EncDeck = []*e.Card{mk("Advance"), mk("False Alarm"), mk("Assault")}
	})
	choose(t, g, "end_turn")
	if g.Pending().Kind != "defend" {
		t.Fatalf("want the defend decision, got %v", keys(g))
	}
	if len(g.S.EncDeck) != 2 || g.S.EncDeck[0].Def.Name != "False Alarm" {
		t.Fatalf("encounter deck %d cards at the defend decision; want the boost card already dealt", len(g.S.EncDeck))
	}
}

func TestSurgeJoinsTheDealtQueue(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Villain.Stunned = true
		s.Dealt = []*e.Card{mk("Caught Off Guard")} // surges: nothing to discard
		s.EncDeck = []*e.Card{mk("Advance"), mk("False Alarm")}
		s.EncDiscard = []*e.Card{mk("Assault")}
	})
	choose(t, g, "end_turn") // the villain phase needs no decision
	advance := slices.Index(g.Log, "Encounter card revealed: Advance (treachery).")
	alarm := slices.Index(g.Log, "Encounter card revealed: False Alarm (treachery).")
	if advance < 0 || alarm < advance {
		t.Fatalf("Advance at %d, False Alarm at %d; the surge card must be revealed after the queued card", advance, alarm)
	}
}

func TestEmptyEncounterDeckAndDiscardLoses(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.Villain.Stunned = true
		s.EncDeck = []*e.Card{mk("Advance")}
		s.EncDiscard = nil
	})
	choose(t, g, "end_turn")
	finishRound(t, g)
	if r := g.Result(); !r.Over || r.Won {
		t.Fatalf("result %+v; want a loss when both encounter piles are empty", r)
	}
}

func TestPlayerDeckResetsWhenADiscardArrives(t *testing.T) {
	g := newGameWith(t, []string{"For Justice!", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		s.MainScheme.Threat = 5
		s.EncDeck = []*e.Card{mk("Advance"), mk("Assault")}
	})
	choose(t, g, "play:For Justice!>The Break-In!|pay:Genius")
	if len(g.S.Deck) != 2 || len(g.S.Discard) != 0 || len(g.S.Dealt) != 1 {
		t.Fatalf("deck %d discard %d dealt %d; want the empty deck reset at once", len(g.S.Deck), len(g.S.Discard), len(g.S.Dealt))
	}
}

func TestSearchEmptyingEncounterDeckResetsIt(t *testing.T) {
	g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
		s.EncDeck = []*e.Card{mk("Breakin' & Takin'")}
		s.EncDiscard = []*e.Card{mk("Advance")}
	})
	d, err := Get("01095")
	if err != nil {
		t.Fatal(err)
	}
	d.Script.OnReveal(g, &e.Card{ID: 99, Def: d}) // finds Breakin' & Takin' in the deck
	g.Run()
	if len(g.S.EncDeck) != 1 || g.S.AccelTokens != 1 {
		t.Fatalf("encounter deck %d tokens %d; want it reset as soon as the search emptied it", len(g.S.EncDeck), g.S.AccelTokens)
	}
}

func TestFirstAidHealsAnyDamagedCharacter(t *testing.T) {
	g := newGameWith(t, []string{"First Aid", "Genius"}, func(s *e.State, mk func(string) *e.Card) {
		dd := mk("Daredevil")
		dd.Damage = 3
		s.Play = append(s.Play, dd)
		s.Villain.Damage = 1
	})
	var targets []string
	for _, k := range keys(g) {
		if rest, ok := strings.CutPrefix(k, "play:First Aid>"); ok {
			targets = append(targets, strings.Split(rest, "|")[0])
		}
	}
	if !slices.Equal(targets, []string{"Daredevil", "Rhino"}) {
		t.Fatalf("First Aid targets %v; want only the damaged characters, enemies included", targets)
	}
	choose(t, g, "play:First Aid>Daredevil|pay:Genius")
	if d := g.S.Play[0].Damage; d != 1 {
		t.Fatalf("Daredevil has %d damage; want 1 after healing 2", d)
	}
}

func TestGreatResponsibilityOutsideVillainPhase(t *testing.T) {
	card := func(ref string) *e.Card {
		d, err := Get(ref)
		if err != nil {
			t.Fatal(err)
		}
		return &e.Card{Def: d}
	}
	s := &e.State{Hero: card("Spider-Man"), Villain: card("01094"), MainScheme: card("The Break-In!"),
		Phase: e.PhasePlayer, Hand: []*e.Card{card("Great Responsibility")}, EncDeck: []*e.Card{card(filler)}}
	g := e.New(s, 1)
	g.PlaceThreatWindowed(s.MainScheme, 2)
	g.Run()
	choose(t, g, "play:Great Responsibility")
	if s.Hero.Damage != 2 || s.MainScheme.Threat != 0 {
		t.Fatalf("hero damage %d threat %d; want the player-phase threat taken as damage", s.Hero.Damage, s.MainScheme.Threat)
	}
}

func TestSweepingSwoopBoostStunsDamagedCharacter(t *testing.T) {
	for _, tough := range []bool{false, true} {
		g := newGameWith(t, nil, func(s *e.State, mk func(string) *e.Card) {
			s.Hero.Tough = tough
			s.EncDeck = []*e.Card{mk("Sweeping Swoop"), mk(filler), mk(filler)}
		})
		choose(t, g, "end_turn")
		choose(t, g, "defend:none")
		if g.S.Hero.Stunned == tough {
			t.Fatalf("tough %v: hero stunned %v; the boost should stun only a damaged hero", tough, g.S.Hero.Stunned)
		}
	}
}

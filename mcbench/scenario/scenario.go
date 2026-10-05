// Package scenario loads Marvel Champions scenarios. A scenario is a hero
// deck plus a villain (with its main scheme and encounter set) and the extra
// encounter sets shuffled in (standard, modular). Every game starts from
// regular setup; the benchmark plays it to the end.
package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mcbench/cards"
	"mcbench/engine"
)

// Scenario is one deck-vs-villain matchup.
type Scenario struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// Deck is the id of a file in decks/.
	Deck string `json:"deck"`
	// Villain is the id of a file in villains/.
	Villain string `json:"villain"`
	// EncounterSets are ids of files in encounter-sets/ (standard, modular).
	EncounterSets []string `json:"encounter_sets"`
	// Seeds fix the shuffles. Every agent plays the same seeds, so the same
	// choices always produce the same game.
	Seeds []uint64 `json:"seeds"`
	// MaxRounds ends an unfinished game (scored as not won).
	MaxRounds int `json:"max_rounds"`
	// MaxDecisions guards against runaway games.
	MaxDecisions int `json:"max_decisions"`
	// Scoring overrides the default criterion weights.
	Scoring map[string]float64 `json:"scoring,omitempty"`

	deck    Deck
	villain Villain
	sets    []EncounterSet
	Path    string `json:"-"`
}

// Deck is a hero's identity, deck, obligation and nemesis set.
type Deck struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Identity   string   `json:"identity"`
	Cards      []string `json:"cards"`
	Obligation string   `json:"obligation"`
	Nemesis    []string `json:"nemesis"`
}

// Villain is a villain's stages (I, II, ...), main scheme and encounter set.
type Villain struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Stages     []string `json:"stages"`
	MainScheme string   `json:"main_scheme"`
	Cards      []string `json:"cards"`
}

// EncounterSet is a standard or modular encounter set.
type EncounterSet struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Cards []string `json:"cards"`
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// Load reads a scenario and the deck, villain and encounter sets it names.
// root is the directory holding decks/, villains/ and encounter-sets/.
func Load(path, root string) (*Scenario, error) {
	var s Scenario
	if err := readJSON(path, &s); err != nil {
		return nil, err
	}
	s.Path = path
	if s.ID == "" {
		s.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if len(s.Seeds) == 0 {
		return nil, fmt.Errorf("%s: no seeds", s.ID)
	}
	if s.MaxRounds == 0 {
		s.MaxRounds = 20
	}
	if s.MaxDecisions == 0 {
		s.MaxDecisions = 1000
	}
	if err := readJSON(filepath.Join(root, "decks", s.Deck+".json"), &s.deck); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(root, "villains", s.Villain+".json"), &s.villain); err != nil {
		return nil, err
	}
	for _, id := range s.EncounterSets {
		var es EncounterSet
		if err := readJSON(filepath.Join(root, "encounter-sets", id+".json"), &es); err != nil {
			return nil, err
		}
		s.sets = append(s.sets, es)
	}
	// Build once to surface unknown cards at load time.
	if _, err := s.NewGame(s.Seeds[0]); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadDir loads every scenario in dir; decks etc. are found under root.
func LoadDir(dir, root string) ([]*Scenario, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []*Scenario
	for _, p := range paths {
		s, err := Load(p, root)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Scenario) DeckName() string    { return s.deck.Name }
func (s *Scenario) VillainName() string { return s.villain.Name }

// NewGame builds the game for a seed, ready for engine.Game.StartGame.
func (s *Scenario) NewGame(seed uint64) (*engine.Game, error) {
	id := 0
	var firstErr error
	mk := func(ref string) *engine.Card {
		d, err := cards.Get(ref)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", s.ID, err)
			}
			d = &engine.CardDef{Code: "?", Name: ref}
		}
		id++
		return &engine.Card{ID: id, Def: d, Counters: d.Uses}
	}
	mkAll := func(refs []string) []*engine.Card {
		out := make([]*engine.Card, 0, len(refs))
		for _, r := range refs {
			out = append(out, mk(r))
		}
		return out
	}
	v := s.villain
	if len(v.Stages) == 0 {
		return nil, fmt.Errorf("%s: villain %s has no stages", s.ID, v.ID)
	}
	st := &engine.State{
		Hero:       mk(s.deck.Identity),
		Deck:       mkAll(s.deck.Cards),
		Villain:    mk(v.Stages[0]),
		MainScheme: mk(v.MainScheme),
		EncDeck:    mkAll(v.Cards),
		Nemesis:    mkAll(s.deck.Nemesis),
	}
	for _, ref := range v.Stages[1:] {
		d, err := cards.Get(ref)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.ID, err)
		}
		st.VillainStages = append(st.VillainStages, d)
	}
	for _, es := range s.sets {
		st.EncDeck = append(st.EncDeck, mkAll(es.Cards)...)
	}
	if s.deck.Obligation != "" {
		st.EncDeck = append(st.EncDeck, mk(s.deck.Obligation))
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if st.Hero.Def.Type != engine.TypeHero || st.Hero.Def.Back == nil {
		return nil, fmt.Errorf("%s: identity %q is not a double-sided hero", s.ID, s.deck.Identity)
	}
	for _, c := range st.Deck {
		if !c.Def.IsPlayerCard() {
			return nil, fmt.Errorf("%s: deck card %q is not a player card", s.ID, c.Def.Name)
		}
	}
	return engine.New(st, seed), nil
}

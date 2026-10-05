// Package scenario loads Marvel Champions scenarios. A scenario is a hero
// deck plus a villain (with its main scheme and encounter set) and the extra
// encounter sets shuffled in (standard, modular). Every game starts from
// regular setup; the benchmark plays it to the end.
package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"mcbench/constants"
	"mcbench/game/cards"
	"mcbench/game/engine"
)

// unknownCardCode marks a placeholder card built when a reference does not
// resolve; building continues so every missing card is reported together.
const unknownCardCode = "?"

func readYAML(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.UnmarshalWithOptions(data, v, yaml.Strict()); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// Load reads a scenario and the deck, villain and encounter sets it names.
// root is the directory holding decks/, villains/ and encounter-sets/.
func Load(path, root string) (*Scenario, error) {
	var s Scenario
	if err := readYAML(path, &s); err != nil {
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
	if err := readYAML(filepath.Join(root, constants.DirDecks, s.Deck+constants.ExtYAML), &s.deck); err != nil {
		return nil, err
	}
	if err := readYAML(filepath.Join(root, constants.DirVillains, s.Villain+constants.ExtYAML), &s.villain); err != nil {
		return nil, err
	}
	for _, id := range s.EncounterSets {
		var es EncounterSet
		if err := readYAML(filepath.Join(root, constants.DirEncounterSets, id+constants.ExtYAML), &es); err != nil {
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
	paths, err := filepath.Glob(filepath.Join(dir, "*"+constants.ExtYAML))
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

// cardMaker creates the cards of one game, numbering them and remembering
// the first card that could not be resolved.
type cardMaker struct {
	scenario string
	nextID   int
	err      error
}

func (m *cardMaker) one(ref string) *engine.Card {
	d, err := cards.Get(ref)
	if err != nil {
		if m.err == nil {
			m.err = fmt.Errorf("%s: %w", m.scenario, err)
		}
		d = &engine.CardDef{Code: unknownCardCode, Name: ref}
	}
	m.nextID++
	return &engine.Card{ID: m.nextID, Def: d, Counters: d.Uses}
}

func (m *cardMaker) many(refs []string) []*engine.Card {
	out := make([]*engine.Card, 0, len(refs))
	for _, r := range refs {
		out = append(out, m.one(r))
	}
	return out
}

// NewGame builds the game for a seed, ready for engine.Game.StartGame.
func (s *Scenario) NewGame(seed uint64) (*engine.Game, error) {
	st, err := s.buildState()
	if err != nil {
		return nil, err
	}
	if err := s.checkState(st); err != nil {
		return nil, err
	}
	return engine.New(st, seed), nil
}

func (s *Scenario) buildState() (*engine.State, error) {
	v := s.villain
	if len(v.Stages) == 0 {
		return nil, fmt.Errorf("%s: villain %s has no stages", s.ID, v.ID)
	}
	m := &cardMaker{scenario: s.ID}
	st := &engine.State{
		Hero:       m.one(s.deck.Identity),
		Deck:       m.many(s.deck.Cards),
		Villain:    m.one(v.Stages[0]),
		MainScheme: m.one(v.MainScheme),
		EncDeck:    m.many(v.Cards),
		Nemesis:    m.many(s.deck.Nemesis),
	}
	for _, ref := range v.Stages[1:] {
		d, err := cards.Get(ref)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.ID, err)
		}
		st.VillainStages = append(st.VillainStages, d)
	}
	for _, es := range s.sets {
		st.EncDeck = append(st.EncDeck, m.many(es.Cards)...)
	}
	if s.deck.Obligation != "" {
		st.EncDeck = append(st.EncDeck, m.one(s.deck.Obligation))
	}
	return st, m.err
}

func (s *Scenario) checkState(st *engine.State) error {
	if st.Hero.Def.Type != engine.TypeHero || st.Hero.Def.Back == nil {
		return fmt.Errorf("%s: identity %q is not a double-sided hero", s.ID, s.deck.Identity)
	}
	for _, c := range st.Deck {
		if !c.Def.IsPlayerCard() {
			return fmt.Errorf("%s: deck card %q is not a player card", s.ID, c.Def.Name)
		}
	}
	return nil
}

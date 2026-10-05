package scenario

// Scenario is one deck-vs-villain matchup.
type Scenario struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	// Deck is the id of a file in decks/.
	Deck string `yaml:"deck"`
	// Villain is the id of a file in villains/.
	Villain string `yaml:"villain"`
	// EncounterSets are ids of files in encounter-sets/ (standard, modular).
	EncounterSets []string `yaml:"encounter_sets"`
	// Seeds fix the shuffles. Every player plays the same seeds, so the same
	// choices always produce the same game.
	Seeds []uint64 `yaml:"seeds"`
	// MaxRounds ends an unfinished game (scored as not won).
	MaxRounds int `yaml:"max_rounds"`
	// MaxDecisions guards against runaway games.
	MaxDecisions int `yaml:"max_decisions"`
	// Scoring overrides the default criterion weights.
	Scoring map[string]float64 `yaml:"scoring"`

	deck    Deck
	villain Villain
	sets    []EncounterSet
	Path    string `yaml:"-"`
}

// Deck is a hero's identity, deck, obligation and nemesis set.
type Deck struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Identity   string   `yaml:"identity"`
	Cards      []string `yaml:"cards"`
	Obligation string   `yaml:"obligation"`
	Nemesis    []string `yaml:"nemesis"`
}

// Villain is a villain's stages (I, II, ...), main scheme and encounter set.
type Villain struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Stages     []string `yaml:"stages"`
	MainScheme string   `yaml:"main_scheme"`
	Cards      []string `yaml:"cards"`
}

// EncounterSet is a standard or modular encounter set.
type EncounterSet struct {
	ID    string   `yaml:"id"`
	Name  string   `yaml:"name"`
	Cards []string `yaml:"cards"`
}

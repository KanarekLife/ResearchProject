package cards

// Card definitions are data: printed stats and a list of abilities. loader.go
// compiles these into engine Script hooks; effects.go interprets them. See
// data/cards/*.yaml and docs/scenarios.md.

// cardFile is one YAML file of card definitions.
type cardFile struct {
	Cards []CardDoc `yaml:"cards"`
}

// CardDoc is one card face as written in a data file.
type CardDoc struct {
	Code      string   `yaml:"code"`
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type"`
	Unique    bool     `yaml:"unique"`
	Aspect    string   `yaml:"aspect"`
	MaxInPlay int      `yaml:"max_in_play"`
	Traits    []string `yaml:"traits"`
	Text      string   `yaml:"text"`

	Cost      int      `yaml:"cost"`
	Resources []string `yaml:"resources"`

	HP       int `yaml:"hp"`
	ATK      int `yaml:"atk"`
	THW      int `yaml:"thw"`
	DEF      int `yaml:"def"`
	REC      int `yaml:"rec"`
	SCH      int `yaml:"sch"`
	HandSize int `yaml:"hand_size"`
	AtkCons  int `yaml:"atk_cons"`
	ThwCons  int `yaml:"thw_cons"`

	Boost int `yaml:"boost"`

	StartingThreat int  `yaml:"starting_threat"`
	TargetThreat   int  `yaml:"target_threat"`
	Escalation     int  `yaml:"escalation"`
	Hazard         int  `yaml:"hazard"`
	Acceleration   int  `yaml:"acceleration"`
	Crisis         bool `yaml:"crisis"`
	// NoThreatRemoval: threat cannot be removed from this scheme.
	NoThreatRemoval bool `yaml:"no_threat_removal"`

	Guard       bool `yaml:"guard"`
	Toughness   bool `yaml:"toughness"`
	Surge       bool `yaml:"surge"`
	Quickstrike bool `yaml:"quickstrike"`
	Uses        int  `yaml:"uses"`
	Retaliate   int  `yaml:"retaliate"`
	// ExtraBoost: the villain gets this many additional boost cards when it
	// attacks.
	ExtraBoost     int  `yaml:"extra_boost"`
	AllyMustDefend bool `yaml:"ally_must_defend"`
	NoThwart       bool `yaml:"no_thwart"`

	AttachATK      int    `yaml:"attach_atk"`
	AttachTHW      int    `yaml:"attach_thw"`
	AttachSCH      int    `yaml:"attach_sch"`
	AttachOverkill bool   `yaml:"attach_overkill"`
	AttachTo       string `yaml:"attach_to"`
	DroneATK       int    `yaml:"drone_atk"`
	DroneHP        int    `yaml:"drone_hp"`

	HeroATK        int      `yaml:"hero_atk"`
	HeroTHW        int      `yaml:"hero_thw"`
	HeroDEF        int      `yaml:"hero_def"`
	HeroHP         int      `yaml:"hero_hp"`
	HeroTraits     []string `yaml:"hero_traits"`
	VillainHP      int      `yaml:"villain_hp"`
	AllyLimit      int      `yaml:"ally_limit"`
	CostFromTarget bool     `yaml:"cost_from_target"`

	DoubleFor string `yaml:"double_for"`

	Back      *CardDoc  `yaml:"back"`
	Abilities []Ability `yaml:"abilities"`
}

// Ability is one trigger plus the effects it runs. Trigger is one of play,
// action, reveal, setup, defeated, interrupt, response, forced, resource,
// stat, boost or boost_flip.
type Ability struct {
	Trigger      string   `yaml:"trigger"`
	On           string   `yaml:"on"`
	When         []string `yaml:"when"`
	Target       string   `yaml:"target"`
	Paid         *Paid    `yaml:"paid"`
	Key          string   `yaml:"key"`
	Text         string   `yaml:"text"`
	Effects      []Effect `yaml:"effects"`
	Gives        string   `yaml:"gives"`
	Usable       []string `yaml:"usable"`
	OncePerRound bool     `yaml:"once_per_round"`
	Stat         string   `yaml:"stat"`
	Bonus        string   `yaml:"bonus"`
	// Max caps a stat ability's stat (bonus included).
	Max int `yaml:"max"`
	// FromDiscard: the resource ability generates the printed resources of
	// the top card of your discard pile.
	FromDiscard bool `yaml:"from_discard"`
}

// Paid is the cost of an activated ability.
type Paid struct {
	Count     int      `yaml:"count"`
	Resources []string `yaml:"resources"`
}

// Effect is one instruction. Verb names the operation; the other fields are
// its parameters. When gates the effect; the nested lists run on a branch.
type Effect struct {
	Verb            string   `yaml:"verb"`
	When            []string `yaml:"when"`
	Amount          int      `yaml:"amount"`
	Damage          int      `yaml:"damage"`
	N               int      `yaml:"n"`
	Target          string   `yaml:"target"`
	Choose          bool     `yaml:"choose"`
	Resource        string   `yaml:"resource"`
	Code            string   `yaml:"code"`
	Result          string   `yaml:"result"`
	Max             int      `yaml:"max"`
	Prompt          string   `yaml:"prompt"`
	DetachWhenEmpty bool     `yaml:"detach_when_empty"`
	Stat            string   `yaml:"stat"`
	Trait           string   `yaml:"trait"`
	// Per multiplies Amount (or Damage) by a count, e.g. drones or counters.
	Per string `yaml:"per"`
	// Overkill gives an attack overkill.
	Overkill bool `yaml:"overkill"`
	// Other: a chosen scheme must differ from the effect's target.
	Other bool `yaml:"other"`
	// Then runs after the effect resolves (a choice is made, an attack or
	// thwart is not replaced by a status).
	Then      []Effect `yaml:"then"`
	IfZero    []Effect `yaml:"if_zero"`
	IfAlready []Effect `yaml:"if_already"`
	IfEmpty   []Effect `yaml:"if_empty"`
	Options   []Choice `yaml:"options"`
}

// Choice is one branch of a choose effect. Paid, if set, is its cost.
type Choice struct {
	Text    string   `yaml:"text"`
	Key     string   `yaml:"key"`
	When    []string `yaml:"when"`
	Paid    *Paid    `yaml:"paid"`
	Effects []Effect `yaml:"effects"`
}

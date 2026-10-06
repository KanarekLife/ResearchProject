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

	Guard       bool `yaml:"guard"`
	Toughness   bool `yaml:"toughness"`
	Surge       bool `yaml:"surge"`
	Quickstrike bool `yaml:"quickstrike"`
	Uses        int  `yaml:"uses"`

	AttachATK      int  `yaml:"attach_atk"`
	AttachSCH      int  `yaml:"attach_sch"`
	AttachOverkill bool `yaml:"attach_overkill"`

	HeroATK int `yaml:"hero_atk"`
	HeroTHW int `yaml:"hero_thw"`
	HeroDEF int `yaml:"hero_def"`

	DoubleFor string `yaml:"double_for"`

	Back      *CardDoc  `yaml:"back"`
	Abilities []Ability `yaml:"abilities"`
}

// Ability is one trigger plus the effects it runs. Trigger is one of play,
// action, reveal, defeated, interrupt, forced, resource, stat or boost.
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
	IfZero          []Effect `yaml:"if_zero"`
	IfAlready       []Effect `yaml:"if_already"`
	IfEmpty         []Effect `yaml:"if_empty"`
	Options         []Choice `yaml:"options"`
}

// Choice is one branch of a choose effect.
type Choice struct {
	Text    string   `yaml:"text"`
	Key     string   `yaml:"key"`
	When    []string `yaml:"when"`
	Effects []Effect `yaml:"effects"`
}

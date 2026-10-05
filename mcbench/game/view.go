package game

import (
	"mcbench/engine"
)

// View is the public game state: everything the player may know. It never
// contains deck order, the encounter deck's contents or face-down cards.
type View struct {
	Round int    `json:"round"`
	Phase string `json:"phase"`

	Hero                Hero     `json:"hero"`
	Hand                []Card   `json:"hand"`
	InPlay              []Card   `json:"in_play"`
	DeckSize            int      `json:"deck_size"`
	Discard             []string `json:"discard"`
	FaceDownEncounters  int      `json:"face_down_encounter_cards"`
	Villain             Enemy    `json:"villain"`
	NextStages          []Enemy  `json:"next_villain_stages"`
	MainScheme          Scheme   `json:"main_scheme"`
	SideSchemes         []Scheme `json:"side_schemes"`
	Minions             []Enemy  `json:"minions"`
	EncounterDeckSize   int      `json:"encounter_deck_size"`
	EncounterDiscard    []string `json:"encounter_discard"`
	AccelerationTokens  int      `json:"acceleration_tokens"`
	NextCardCostReduced int      `json:"next_card_cost_reduction,omitempty"`
}

type Hero struct {
	Name         string   `json:"name"`
	Form         string   `json:"form"` // "hero" or "alter-ego"
	HP           int      `json:"hp"`
	MaxHP        int      `json:"max_hp"`
	ATK          int      `json:"atk,omitempty"`
	THW          int      `json:"thw,omitempty"`
	DEF          int      `json:"def,omitempty"`
	REC          int      `json:"rec,omitempty"`
	HandSize     int      `json:"hand_size"`
	Exhausted    bool     `json:"exhausted"`
	Statuses     []string `json:"statuses,omitempty"`
	FormChanged  bool     `json:"changed_form_this_turn"`
	Ability      string   `json:"ability,omitempty"`
	OtherForm    string   `json:"other_form"`
	OtherAbility string   `json:"other_form_ability,omitempty"`
	Attachments  []Card   `json:"attachments,omitempty"`
}

// Card is a player card in hand or in play.
type Card struct {
	Name      string   `json:"name"`
	Label     string   `json:"label,omitempty"` // "Name#ID" when names repeat in play
	Type      string   `json:"type"`
	Cost      *int     `json:"cost,omitempty"`
	Resources []string `json:"resources,omitempty"`
	HP        int      `json:"hp,omitempty"`
	MaxHP     int      `json:"max_hp,omitempty"`
	ATK       int      `json:"atk,omitempty"`
	THW       int      `json:"thw,omitempty"`
	AtkCons   int      `json:"atk_consequential,omitempty"`
	ThwCons   int      `json:"thw_consequential,omitempty"`
	Counters  int      `json:"counters,omitempty"`
	Exhausted bool     `json:"exhausted,omitempty"`
	Statuses  []string `json:"statuses,omitempty"`
	Text      string   `json:"text,omitempty"`
}

// Enemy is the villain (or a villain stage) or a minion.
type Enemy struct {
	Name        string   `json:"name"`
	Label       string   `json:"label,omitempty"`
	Stage       int      `json:"stage,omitempty"`
	HP          int      `json:"hp"`
	MaxHP       int      `json:"max_hp"`
	ATK         int      `json:"atk"`
	SCH         int      `json:"sch"`
	Keywords    []string `json:"keywords,omitempty"`
	Statuses    []string `json:"statuses,omitempty"`
	Text        string   `json:"text,omitempty"`
	Attachments []Card   `json:"attachments,omitempty"`
}

type Scheme struct {
	Name        string   `json:"name"`
	Label       string   `json:"label,omitempty"`
	Threat      int      `json:"threat"`
	Threshold   int      `json:"threshold,omitempty"` // main scheme: you lose when threat reaches it
	PerRound    int      `json:"threat_per_round,omitempty"`
	Icons       []string `json:"icons,omitempty"`
	Text        string   `json:"text,omitempty"`
	Attachments []Card   `json:"attachments,omitempty"`
}

func statuses(c *engine.Card) []string {
	var s []string
	if c.Stunned {
		s = append(s, "stunned")
	}
	if c.Confused {
		s = append(s, "confused")
	}
	if c.Tough {
		s = append(s, "tough")
	}
	return s
}

func resources(rs []engine.Resource) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = string(r)
	}
	return out
}

func names(cs []*engine.Card) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Def.Name
	}
	return out
}

func handCard(c *engine.Card) Card {
	d := c.Def
	out := Card{Name: d.Name, Type: string(d.Type), Resources: resources(d.Resources), Text: d.Text}
	if d.Type != engine.TypeResource {
		cost := d.Cost
		out.Cost = &cost
	}
	if d.Type == engine.TypeAlly {
		out.HP, out.MaxHP, out.ATK, out.THW, out.AtkCons, out.ThwCons = d.HP, d.HP, d.ATK, d.THW, d.AtkCons, d.ThwCons
	}
	return out
}

func playCard(g *engine.Game, c *engine.Card) Card {
	d := c.Face()
	out := Card{Name: d.Name, Label: g.Label(c), Type: string(d.Type), Counters: c.Counters, Exhausted: c.Exhausted, Statuses: statuses(c), Text: d.Text}
	if out.Label == out.Name {
		out.Label = ""
	}
	if d.Type == engine.TypeAlly {
		out.HP, out.MaxHP, out.ATK, out.THW, out.AtkCons, out.ThwCons = c.RemainingHP(), d.HP, g.AllyATK(c), g.AllyTHW(c), d.AtkCons, d.ThwCons
	}
	return out
}

func playCards(g *engine.Game, cs []*engine.Card) []Card {
	out := []Card{}
	for _, c := range cs {
		out = append(out, playCard(g, c))
	}
	return out
}

func enemy(g *engine.Game, c *engine.Card) Enemy {
	d := c.Face()
	e := Enemy{Name: d.Name, Label: g.Label(c), HP: c.RemainingHP(), MaxHP: d.HP, ATK: g.EnemyATK(c), SCH: g.EnemySCH(c),
		Statuses: statuses(c), Text: d.Text, Attachments: playCards(g, c.Attached)}
	if e.Label == e.Name {
		e.Label = ""
	}
	if d.Guard {
		e.Keywords = append(e.Keywords, "guard")
	}
	if d.Quickstrike {
		e.Keywords = append(e.Keywords, "quickstrike")
	}
	return e
}

func scheme(g *engine.Game, c *engine.Card) Scheme {
	d := c.Face()
	s := Scheme{Name: d.Name, Label: g.Label(c), Threat: c.Threat, Text: d.Text, Attachments: playCards(g, c.Attached)}
	if s.Label == s.Name {
		s.Label = ""
	}
	if d.Crisis {
		s.Icons = append(s.Icons, "crisis")
	}
	if d.Hazard > 0 {
		s.Icons = append(s.Icons, "hazard")
	}
	if d.Acceleration > 0 {
		s.Icons = append(s.Icons, "acceleration")
	}
	return s
}

// buildView reads the public state out of the engine.
func buildView(g *engine.Game, totalStages int) View {
	s := g.S
	h := s.Hero
	f := h.Face()
	other := h.Def.Back
	if h.Flipped {
		other = h.Def
	}
	v := View{
		Round: s.Round, Phase: string(s.Phase),
		Hero: Hero{
			Name: f.Name, Form: "hero", HP: h.RemainingHP(), MaxHP: f.HP, HandSize: f.HandSize,
			Exhausted: h.Exhausted, Statuses: statuses(h), FormChanged: s.FormChanged, Ability: f.Text,
			OtherForm: other.Name, OtherAbility: other.Text, Attachments: playCards(g, h.Attached),
		},
		Hand:                []Card{},
		InPlay:              playCards(g, s.Play),
		DeckSize:            len(s.Deck),
		Discard:             names(s.Discard),
		FaceDownEncounters:  len(s.Dealt),
		Villain:             enemy(g, s.Villain),
		MainScheme:          scheme(g, s.MainScheme),
		SideSchemes:         []Scheme{},
		Minions:             []Enemy{},
		EncounterDeckSize:   len(s.EncDeck),
		EncounterDiscard:    names(s.EncDiscard),
		AccelerationTokens:  s.AccelTokens,
		NextCardCostReduced: s.CostReduction,
	}
	if g.IsHero() {
		v.Hero.ATK, v.Hero.THW, v.Hero.DEF = g.HeroATK(), g.HeroTHW(), g.HeroDEF()
	} else {
		v.Hero.Form, v.Hero.REC = "alter-ego", f.REC
	}
	for _, c := range s.Hand {
		v.Hand = append(v.Hand, handCard(c))
	}
	stage := totalStages - len(s.VillainStages)
	v.Villain.Stage = stage
	for i, st := range s.VillainStages {
		v.NextStages = append(v.NextStages, Enemy{Name: st.Name, Stage: stage + 1 + i, HP: st.HP, MaxHP: st.HP, ATK: st.ATK, SCH: st.SCH, Text: st.Text})
	}
	v.MainScheme.Threshold = s.MainScheme.Face().TargetThreat
	v.MainScheme.PerRound = s.MainScheme.Face().Escalation
	for _, c := range s.SideSchemes {
		v.SideSchemes = append(v.SideSchemes, scheme(g, c))
	}
	for _, c := range s.Minions {
		v.Minions = append(v.Minions, enemy(g, c))
	}
	return v
}

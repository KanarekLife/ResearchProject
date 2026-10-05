package session

import (
	"mcbench/constants"
	"mcbench/game/engine"
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

// Hero is the identity and its current form.
type Hero struct {
	Name         string   `json:"name"`
	Form         string   `json:"form"` // hero or alter-ego
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

// Scheme is the main scheme or a side scheme.
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

func buildView(g *engine.Game, totalStages int) View {
	s := g.S
	v := View{
		Round:               s.Round,
		Phase:               string(s.Phase),
		Hero:                heroView(g),
		Hand:                mapCards(g.S.Hand, handCard),
		InPlay:              mapCards(s.Play, playFn(g)),
		DeckSize:            len(s.Deck),
		Discard:             names(s.Discard),
		FaceDownEncounters:  len(s.Dealt),
		Villain:             enemy(g, s.Villain),
		MainScheme:          scheme(g, s.MainScheme),
		SideSchemes:         mapCards(s.SideSchemes, func(c *engine.Card) Scheme { return scheme(g, c) }),
		Minions:             mapCards(s.Minions, func(c *engine.Card) Enemy { return enemy(g, c) }),
		EncounterDeckSize:   len(s.EncDeck),
		EncounterDiscard:    names(s.EncDiscard),
		AccelerationTokens:  s.AccelTokens,
		NextCardCostReduced: s.CostReduction,
	}
	v.Villain.Stage = totalStages - len(s.VillainStages)
	v.NextStages = nextStages(s, v.Villain.Stage)
	v.MainScheme.Threshold = s.MainScheme.Face().TargetThreat
	v.MainScheme.PerRound = s.MainScheme.Face().Escalation
	return v
}

func heroView(g *engine.Game) Hero {
	s := g.S
	h := s.Hero
	face := h.Face()
	other := h.Def.Back
	if h.Flipped {
		other = h.Def
	}
	hero := Hero{
		Name: face.Name, Form: constants.TypeHero, HP: h.RemainingHP(), MaxHP: face.HP,
		HandSize: face.HandSize, Exhausted: h.Exhausted, Statuses: statuses(h),
		FormChanged: s.FormChanged, Ability: face.Text,
		OtherForm: other.Name, OtherAbility: other.Text, Attachments: mapCards(h.Attached, playFn(g)),
	}
	if g.IsHero() {
		hero.ATK, hero.THW, hero.DEF = g.HeroATK(), g.HeroTHW(), g.HeroDEF()
	} else {
		hero.Form, hero.REC = constants.TypeAlterEgo, face.REC
	}
	return hero
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
	out := Card{Name: d.Name, Label: labelOf(g, c), Type: string(d.Type), Counters: c.Counters, Exhausted: c.Exhausted, Statuses: statuses(c), Text: d.Text}
	if d.Type == engine.TypeAlly {
		out.HP, out.MaxHP, out.ATK, out.THW, out.AtkCons, out.ThwCons = c.RemainingHP(), d.HP, g.AllyATK(c), g.AllyTHW(c), d.AtkCons, d.ThwCons
	}
	return out
}

func enemy(g *engine.Game, c *engine.Card) Enemy {
	d := c.Face()
	e := Enemy{
		Name: d.Name, Label: labelOf(g, c), HP: c.RemainingHP(), MaxHP: d.HP,
		ATK: g.EnemyATK(c), SCH: g.EnemySCH(c), Statuses: statuses(c),
		Text: d.Text, Attachments: mapCards(c.Attached, playFn(g)),
	}
	if d.Guard {
		e.Keywords = append(e.Keywords, constants.KeywordGuard)
	}
	if d.Quickstrike {
		e.Keywords = append(e.Keywords, constants.KeywordQuickstrike)
	}
	return e
}

func scheme(g *engine.Game, c *engine.Card) Scheme {
	d := c.Face()
	s := Scheme{Name: d.Name, Label: labelOf(g, c), Threat: c.Threat, Text: d.Text, Attachments: mapCards(c.Attached, playFn(g))}
	if d.Crisis {
		s.Icons = append(s.Icons, constants.IconCrisis)
	}
	if d.Hazard > 0 {
		s.Icons = append(s.Icons, constants.IconHazard)
	}
	if d.Acceleration > 0 {
		s.Icons = append(s.Icons, constants.IconAcceleration)
	}
	return s
}

func nextStages(s *engine.State, current int) []Enemy {
	var out []Enemy
	for i, st := range s.VillainStages {
		out = append(out, Enemy{Name: st.Name, Stage: current + 1 + i, HP: st.HP, MaxHP: st.HP, ATK: st.ATK, SCH: st.SCH, Text: st.Text})
	}
	return out
}

func statuses(c *engine.Card) []string {
	var s []string
	if c.Stunned {
		s = append(s, constants.StatusStunned)
	}
	if c.Confused {
		s = append(s, constants.StatusConfused)
	}
	if c.Tough {
		s = append(s, constants.StatusTough)
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

// mapCards converts a zone to its view form; the result is never nil, so it
// encodes as [] rather than null.
func mapCards[T any](cs []*engine.Card, f func(*engine.Card) T) []T {
	out := make([]T, 0, len(cs))
	for _, c := range cs {
		out = append(out, f(c))
	}
	return out
}

func playFn(g *engine.Game) func(*engine.Card) Card {
	return func(c *engine.Card) Card { return playCard(g, c) }
}

// labelOf is the card's unique label, or "" when it is just its name.
func labelOf(g *engine.Game, c *engine.Card) string {
	if l := g.Label(c); l != c.Face().Name {
		return l
	}
	return ""
}

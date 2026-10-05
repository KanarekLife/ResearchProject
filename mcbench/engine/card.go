package engine

import "strings"

// CardType is the printed card type.
type CardType string

const (
	TypeHero       CardType = "hero"
	TypeAlterEgo   CardType = "alter-ego"
	TypeAlly       CardType = "ally"
	TypeEvent      CardType = "event"
	TypeSupport    CardType = "support"
	TypeUpgrade    CardType = "upgrade"
	TypeResource   CardType = "resource"
	TypeVillain    CardType = "villain"
	TypeMainScheme CardType = "main scheme"
	TypeSideScheme CardType = "side scheme"
	TypeMinion     CardType = "minion"
	TypeTreachery  CardType = "treachery"
	TypeAttachment CardType = "attachment"
	TypeObligation CardType = "obligation"
)

// Resource is a resource symbol.
type Resource string

const (
	Physical Resource = "physical"
	Mental   Resource = "mental"
	Energy   Resource = "energy"
	Wild     Resource = "wild"
)

// CardDef is the static definition of one card face. Numbers are for a single
// player (solo); per-player values ("*") are already multiplied out.
type CardDef struct {
	Code   string
	Name   string
	Type   CardType
	Unique bool
	// Aspect is the player card class: "Hero", "Justice", "Basic", ...
	Aspect string
	// MaxInPlay limits copies you may control (e.g. "Max 1 per player").
	MaxInPlay int
	Traits    []string
	// Text is our own short paraphrase of the card's ability, shown to agents.
	Text string

	Cost      int
	Resources []Resource

	HP, ATK, THW, DEF, REC, SCH, HandSize int
	// Consequential damage an ally takes after attacking / thwarting.
	AtkCons, ThwCons int

	Boost int // number of boost icons

	StartingThreat, TargetThreat, Escalation int
	Hazard, Acceleration                     int
	Crisis                                   bool

	Guard, Toughness, Surge, Quickstrike bool
	Uses                                 int // counters placed when entering play

	// Modifiers granted while attached to an enemy.
	AttachATK, AttachSCH int
	// AttachOverkill gives the attached enemy's attacks overkill.
	AttachOverkill bool

	// Bonuses to your hero's stats while this card is in play under your control.
	HeroATK, HeroTHW, HeroDEF int

	// DoubleFor: this resource card generates twice its resources when
	// paying for a card of this aspect (The Power of ...).
	DoubleFor string

	// Back is the other side of a double-sided identity card.
	Back *CardDef

	Script *Script
}

func (d *CardDef) IsPlayerCard() bool {
	switch d.Type {
	case TypeAlly, TypeEvent, TypeSupport, TypeUpgrade, TypeResource:
		return true
	}
	return false
}

func (d *CardDef) HasTrait(t string) bool {
	for _, x := range d.Traits {
		if strings.EqualFold(x, t) {
			return true
		}
	}
	return false
}

// Card is one physical card instance in the game.
type Card struct {
	ID      int
	Def     *CardDef
	Flipped bool // identity is showing its Back face

	Exhausted bool
	Damage    int
	Threat    int
	Counters  int

	Stunned, Confused, Tough bool

	Attached []*Card
}

// Face is the currently visible face.
func (c *Card) Face() *CardDef {
	if c.Flipped && c.Def.Back != nil {
		return c.Def.Back
	}
	return c.Def
}

func (c *Card) Name() string { return c.Face().Name }

func (c *Card) RemainingHP() int { return c.Face().HP - c.Damage }

// Script holds the scripted behavior of a card. Every hook is optional.
type Script struct {
	// Play: events resolve, other cards enter play. Targets lists legal
	// targets; nil Targets means the card is played without a target.
	CanPlay func(g *Game, c *Card) bool
	Targets func(g *Game, c *Card) []*Card
	OnPlay  func(g *Game, c *Card, target *Card)
	// PlayWindow is the timing window an event can be played in. Empty means
	// the player's turn (an "Action").
	PlayWindow Trigger
	// PlayIf filters the window event for interrupt/response events.
	PlayIf func(g *Game, c *Card, ev *Event) bool
	// OnPlayEvent resolves an interrupt/response event against its trigger.
	OnPlayEvent func(g *Game, c *Card, ev *Event)

	// OnReveal is an encounter card's "When Revealed" effect.
	OnReveal func(g *Game, c *Card)
	// OnDefeated runs when a side scheme is defeated.
	OnDefeated func(g *Game, c *Card)

	// Forced abilities of a card in play.
	Forced map[Trigger]func(g *Game, c *Card, ev *Event)

	// Actions lists activated abilities usable on the player's turn.
	Actions func(g *Game, c *Card) []Option

	// Resource ability of a card in play ("generate a resource").
	ResourceAbility *ResourceAbility

	// StatBonus adds to this card's own ATK/THW ("atk", "thw") while in play.
	StatBonus func(g *Game, c *Card, stat string) int
}

// ResourceAbility generates a resource from a card in play.
type ResourceAbility struct {
	Gives  Resource
	Usable func(g *Game, c *Card) bool
	Use    func(g *Game, c *Card)
}

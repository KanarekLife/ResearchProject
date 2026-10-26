package engine

import (
	"strings"

	"mcbench/constants"
)

// CardType is the printed card type.
type CardType string

const (
	TypeHero        CardType = constants.TypeHero
	TypeAlterEgo    CardType = constants.TypeAlterEgo
	TypeAlly        CardType = constants.TypeAlly
	TypeEvent       CardType = constants.TypeEvent
	TypeSupport     CardType = constants.TypeSupport
	TypeUpgrade     CardType = constants.TypeUpgrade
	TypeResource    CardType = constants.TypeResource
	TypeVillain     CardType = constants.TypeVillain
	TypeMainScheme  CardType = constants.TypeMainScheme
	TypeSideScheme  CardType = constants.TypeSideScheme
	TypeMinion      CardType = constants.TypeMinion
	TypeTreachery   CardType = constants.TypeTreachery
	TypeAttachment  CardType = constants.TypeAttachment
	TypeObligation  CardType = constants.TypeObligation
	TypeEnvironment CardType = constants.TypeEnvironment
)

// Resource is a resource symbol.
type Resource string

const (
	Physical Resource = constants.ResourcePhysical
	Mental   Resource = constants.ResourceMental
	Energy   Resource = constants.ResourceEnergy
	Wild     Resource = constants.ResourceWild
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
	// NoThreatRemoval: threat cannot be removed from this scheme.
	NoThreatRemoval bool

	Guard, Toughness, Surge, Quickstrike bool
	Uses                                 int // counters placed when entering play
	// Retaliate deals this much damage to whoever attacks the character
	// (on an attachment: to whoever attacks its host).
	Retaliate int
	// ExtraBoost is the number of additional boost cards the villain gets
	// when it attacks.
	ExtraBoost int
	// AllyMustDefend: you must defend this enemy's attacks with an ally if able.
	AllyMustDefend bool
	// NoThwart: the ally has no THW and cannot thwart.
	NoThwart bool

	// Modifiers granted to the card this one is attached to.
	AttachATK, AttachTHW, AttachSCH int
	// AttachOverkill gives the attached enemy's attacks overkill.
	AttachOverkill bool
	// AttachTo names where an encounter attachment goes: the villain (empty)
	// or the environment.
	AttachTo string
	// DroneATK and DroneHP are added to each facedown Drone minion while this
	// card is in play.
	DroneATK, DroneHP int

	// Bonuses to your hero while this card is in play under your control.
	HeroATK, HeroTHW, HeroDEF, HeroHP int
	// HeroTraits are traits your hero gains while this card is in play.
	HeroTraits []string
	// VillainHP is added to the villain's hit points while this card is in play.
	VillainHP int
	// AllyLimit raises your ally limit while this card is in play.
	AllyLimit int
	// CostFromTarget: playing the card also costs its target's printed cost
	// (Make the Call).
	CostFromTarget bool

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

	// Facedown cards attached to a card in play are themselves out of play.
	Facedown bool
	// Drone, when set, is the face of a player card put into play facedown as
	// a Drone minion; the card itself stays hidden.
	Drone *CardDef
	// TempATK and TempTHW last until the end of the phase.
	TempATK, TempTHW int

	Attached []*Card
}

// Face is the currently visible face.
func (c *Card) Face() *CardDef {
	if c.Drone != nil {
		return c.Drone
	}
	if c.Flipped && c.Def.Back != nil {
		return c.Def.Back
	}
	return c.Def
}

func (c *Card) Name() string { return c.Face().Name }

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
	// OnSetup is a main scheme's "Setup" effect, resolved during setup.
	OnSetup func(g *Game, c *Card)
	// OnDefeated runs when a side scheme is defeated.
	OnDefeated func(g *Game, c *Card)
	// BoostOnDamaged is a boost card's "Boost" ability: it runs for each
	// friendly character damaged by the activation the card boosts.
	BoostOnDamaged func(g *Game, c, damaged *Card)
	// OnBoost is a boost card's "Boost" ability that resolves as the card is
	// flipped (ev.Attack tells an attack from a scheme; ev.Target is the
	// defender, nil when undefended).
	OnBoost func(g *Game, c *Card, ev *Event)

	// Forced abilities of a card in play.
	Forced map[Trigger]func(g *Game, c *Card, ev *Event)

	// Actions lists activated abilities usable on the player's turn.
	Actions func(g *Game, c *Card) []Option

	// Windows are the optional interrupts and responses of a card in play,
	// offered in the timing window of their trigger.
	Windows []WindowAbility

	// Resource ability of a card in play ("generate a resource").
	ResourceAbility *ResourceAbility

	// StatBonus adds to this card's own ATK/THW ("atk", "thw") while in play.
	StatBonus func(g *Game, c *Card, stat string) int
}

// WindowAbility is an optional interrupt or response of a card in play.
type WindowAbility struct {
	Trigger   Trigger
	Key, Text string
	If        func(g *Game, c *Card, ev *Event) bool
	Do        func(g *Game, c *Card, ev *Event)
}

// ResourceAbility generates a resource from a card in play.
type ResourceAbility struct {
	Gives Resource
	// FromDiscard generates the printed resources of the top card of your
	// discard pile instead of Gives.
	FromDiscard bool
	Usable      func(g *Game, c *Card) bool
	Use         func(g *Game, c *Card)
}

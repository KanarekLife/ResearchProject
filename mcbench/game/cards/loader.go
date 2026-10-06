package cards

import (
	"fmt"

	"github.com/goccy/go-yaml"

	"mcbench/constants"
	"mcbench/data"
	e "mcbench/game/engine"
)

func init() {
	entries, err := data.FS.ReadDir(constants.EncounterCardDir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		raw, err := data.FS.ReadFile(constants.EncounterCardDir + "/" + entry.Name())
		if err != nil {
			panic(err)
		}
		var file cardFile
		if err := yaml.UnmarshalWithOptions(raw, &file, yaml.Strict()); err != nil {
			panic(fmt.Sprintf("cards: %s: %v", entry.Name(), err))
		}
		for i := range file.Cards {
			if err := validate(&file.Cards[i]); err != nil {
				panic(fmt.Sprintf("cards: %s: %v", entry.Name(), err))
			}
			add(compile(&file.Cards[i]))
		}
	}
}

// compile turns one card document into an engine CardDef with a Script.
func compile(d *CardDoc) *e.CardDef {
	def := &e.CardDef{
		Code: d.Code, Name: d.Name, Type: cardType(d.Type), Unique: d.Unique,
		Aspect: d.Aspect, MaxInPlay: d.MaxInPlay, Traits: d.Traits, Text: d.Text,
		Cost: d.Cost, Resources: resources(d.Resources),
		HP: d.HP, ATK: d.ATK, THW: d.THW, DEF: d.DEF, REC: d.REC, SCH: d.SCH, HandSize: d.HandSize,
		AtkCons: d.AtkCons, ThwCons: d.ThwCons, Boost: d.Boost,
		StartingThreat: d.StartingThreat, TargetThreat: d.TargetThreat, Escalation: d.Escalation,
		Hazard: d.Hazard, Acceleration: d.Acceleration, Crisis: d.Crisis,
		Guard: d.Guard, Toughness: d.Toughness, Surge: d.Surge, Quickstrike: d.Quickstrike, Uses: d.Uses,
		AttachATK: d.AttachATK, AttachSCH: d.AttachSCH, AttachOverkill: d.AttachOverkill,
		HeroATK: d.HeroATK, HeroTHW: d.HeroTHW, HeroDEF: d.HeroDEF,
		DoubleFor: d.DoubleFor,
	}
	if d.Back != nil {
		def.Back = compile(d.Back)
	}
	sc := &e.Script{}
	var actions []Ability
	for _, a := range d.Abilities {
		if a.Trigger == constants.AbilityAction {
			actions = append(actions, a)
		} else {
			applyAbility(sc, d, a)
		}
	}
	if len(actions) > 0 {
		sc.Actions = actionOptions(actions)
	}
	if hasScript(sc) {
		def.Script = sc
	}
	return def
}

func applyAbility(sc *e.Script, d *CardDoc, a Ability) {
	switch a.Trigger {
	case constants.AbilityPlay:
		applyPlay(sc, a)
	case constants.AbilityReveal:
		applyReveal(sc, a)
	case constants.AbilityDefeated:
		effs := a.Effects
		sc.OnDefeated = func(g *e.Game, c *e.Card) { run(g, c, nil, nil, effs) }
	case constants.AbilityInterrupt:
		applyInterrupt(sc, a)
	case constants.AbilityForced:
		applyForced(sc, a)
	case constants.AbilityResource:
		applyResource(sc, d.Code, a)
	case constants.AbilityStat:
		applyStat(sc, a)
	case constants.AbilityBoost:
		effs := a.Effects
		sc.BoostOnDamaged = func(g *e.Game, c, damaged *e.Card) { run(g, c, &e.Event{Target: damaged}, nil, effs) }
	default:
		panic("cards: unknown ability trigger " + a.Trigger)
	}
}

func applyPlay(sc *e.Script, a Ability) {
	effs, whens := a.Effects, a.When
	if len(whens) > 0 {
		sc.CanPlay = func(g *e.Game, c *e.Card) bool { return when(g, c, nil, whens) }
	}
	if a.Target != "" {
		sel := a.Target
		sc.Targets = func(g *e.Game, c *e.Card) []*e.Card { return selectTargets(g, sel) }
	}
	sc.OnPlay = func(g *e.Game, c *e.Card, target *e.Card) { run(g, c, nil, target, effs) }
}

func applyReveal(sc *e.Script, a Ability) {
	effs, whens := a.Effects, a.When
	sc.OnReveal = func(g *e.Game, c *e.Card) {
		if when(g, c, nil, whens) {
			run(g, c, nil, nil, effs)
		}
	}
}

func applyInterrupt(sc *e.Script, a Ability) {
	effs, whens := a.Effects, a.When
	sc.PlayWindow = trigger(a.On)
	sc.PlayIf = func(g *e.Game, c *e.Card, ev *e.Event) bool { return when(g, c, ev, whens) }
	sc.OnPlayEvent = func(g *e.Game, c *e.Card, ev *e.Event) { run(g, c, ev, nil, effs) }
}

func applyForced(sc *e.Script, a Ability) {
	effs, whens := a.Effects, a.When
	if sc.Forced == nil {
		sc.Forced = map[e.Trigger]func(*e.Game, *e.Card, *e.Event){}
	}
	sc.Forced[trigger(a.On)] = func(g *e.Game, c *e.Card, ev *e.Event) {
		if when(g, c, ev, whens) {
			run(g, c, ev, nil, effs)
		}
	}
}

func applyResource(sc *e.Script, code string, a Ability) {
	usable, use, once := a.Usable, a.Effects, a.OncePerRound
	key := constants.PrefixResource + code
	sc.ResourceAbility = &e.ResourceAbility{
		Gives: resource(a.Gives),
		Usable: func(g *e.Game, c *e.Card) bool {
			if once && g.S.OncePerRound[key] {
				return false
			}
			return when(g, c, nil, usable)
		},
		Use: func(g *e.Game, c *e.Card) {
			if once {
				g.S.OncePerRound[key] = true
			}
			run(g, c, nil, nil, use)
		},
	}
}

func applyStat(sc *e.Script, a Ability) {
	stat, bonus := a.Stat, a.Bonus
	sc.StatBonus = func(g *e.Game, c *e.Card, s string) int {
		if s != stat {
			return 0
		}
		return evalBonus(g, bonus)
	}
}

func actionOptions(abilities []Ability) func(*e.Game, *e.Card) []e.Option {
	return func(g *e.Game, c *e.Card) []e.Option {
		var opts []e.Option
		for _, a := range abilities {
			if !when(g, c, nil, a.When) {
				continue
			}
			a := a
			do := func() { run(g, c, nil, nil, a.Effects) }
			if a.Paid != nil {
				opts = append(opts, g.NewActionPaid(c, a.Key, a.Text, a.Paid.Count, resources(a.Paid.Resources), do)...)
			} else {
				opts = append(opts, e.NewAction(a.Key, a.Text, do))
			}
		}
		return opts
	}
}

func hasScript(sc *e.Script) bool {
	return sc.CanPlay != nil || sc.Targets != nil || sc.OnPlay != nil ||
		sc.OnReveal != nil || sc.OnDefeated != nil || sc.BoostOnDamaged != nil || sc.PlayWindow != "" ||
		sc.OnPlayEvent != nil || len(sc.Forced) > 0 || sc.Actions != nil ||
		sc.ResourceAbility != nil || sc.StatBonus != nil
}

var cardTypes = map[string]e.CardType{
	constants.TypeHero: e.TypeHero, constants.TypeAlterEgo: e.TypeAlterEgo, constants.TypeAlly: e.TypeAlly,
	constants.TypeEvent: e.TypeEvent, constants.TypeSupport: e.TypeSupport, constants.TypeUpgrade: e.TypeUpgrade,
	constants.TypeResource: e.TypeResource, constants.TypeVillain: e.TypeVillain, constants.TypeMainScheme: e.TypeMainScheme,
	constants.TypeSideScheme: e.TypeSideScheme, constants.TypeMinion: e.TypeMinion, constants.TypeTreachery: e.TypeTreachery,
	constants.TypeAttachment: e.TypeAttachment, constants.TypeObligation: e.TypeObligation,
}

var resourceNames = map[string]e.Resource{
	constants.ResourcePhysical: e.Physical, constants.ResourceMental: e.Mental,
	constants.ResourceEnergy: e.Energy, constants.ResourceWild: e.Wild,
}

var triggers = map[string]e.Trigger{
	constants.TrigEnemyWouldAttack:      e.TrigEnemyWouldAttack,
	constants.TrigVillainAttacks:        e.TrigVillainAttacks,
	constants.TrigWouldTakeAttackDamage: e.TrigWouldTakeAttackDamage,
	constants.TrigEnemyWouldTakeDamage:  e.TrigEnemyWouldTakeDamage,
	constants.TrigTreacheryRevealed:     e.TrigTreacheryRevealed,
	constants.TrigMinionDefeated:        e.TrigMinionDefeated,
	constants.TrigVillainSchemes:        e.TrigVillainSchemes,
	constants.TrigThreatWouldBePlaced:   e.TrigThreatWouldBePlaced,
	constants.TrigThwarted:              e.TrigThwarted,
	constants.TrigRoundEnd:              e.TrigRoundEnd,
	constants.TrigAttackEnded:           e.TrigAttackEnded,
}

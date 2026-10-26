package cards

import (
	"fmt"
	"strconv"

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
		Hazard: d.Hazard, Acceleration: d.Acceleration, Crisis: d.Crisis, NoThreatRemoval: d.NoThreatRemoval,
		Guard: d.Guard, Toughness: d.Toughness, Surge: d.Surge, Quickstrike: d.Quickstrike, Uses: d.Uses,
		Retaliate: d.Retaliate, ExtraBoost: d.ExtraBoost, AllyMustDefend: d.AllyMustDefend, NoThwart: d.NoThwart,
		AttachATK: d.AttachATK, AttachTHW: d.AttachTHW, AttachSCH: d.AttachSCH, AttachOverkill: d.AttachOverkill,
		AttachTo: d.AttachTo, DroneATK: d.DroneATK, DroneHP: d.DroneHP,
		HeroATK: d.HeroATK, HeroTHW: d.HeroTHW, HeroDEF: d.HeroDEF, HeroHP: d.HeroHP, HeroTraits: d.HeroTraits,
		VillainHP: d.VillainHP, AllyLimit: d.AllyLimit, CostFromTarget: d.CostFromTarget,
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
	case constants.AbilitySetup:
		effs := a.Effects
		sc.OnSetup = func(g *e.Game, c *e.Card) { run(g, c, nil, nil, effs) }
	case constants.AbilityDefeated:
		effs := a.Effects
		sc.OnDefeated = func(g *e.Game, c *e.Card) { run(g, c, nil, nil, effs) }
	case constants.AbilityInterrupt, constants.AbilityResponse:
		if d.Type == constants.TypeEvent {
			applyInterrupt(sc, a)
		} else {
			applyWindow(sc, a)
		}
	case constants.AbilityForced:
		applyForced(sc, a)
	case constants.AbilityResource:
		applyResource(sc, d.Code, a)
	case constants.AbilityStat:
		applyStat(sc, d, a)
	case constants.AbilityBoost:
		effs, whens := a.Effects, a.When
		sc.BoostOnDamaged = func(g *e.Game, c, damaged *e.Card) {
			ev := &e.Event{Target: damaged}
			if when(g, c, ev, whens) {
				run(g, c, ev, nil, effs)
			}
		}
	case constants.AbilityBoostFlip:
		effs, whens := a.Effects, a.When
		sc.OnBoost = func(g *e.Game, c *e.Card, ev *e.Event) {
			if when(g, c, ev, whens) {
				run(g, c, ev, nil, effs)
			}
		}
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
		sc.Targets = func(g *e.Game, c *e.Card) []*e.Card { return selectTargets(g, c, sel) }
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

// applyWindow adds an optional interrupt or response of a card in play.
func applyWindow(sc *e.Script, a Ability) {
	effs, whens := a.Effects, a.When
	sc.Windows = append(sc.Windows, e.WindowAbility{
		Trigger: trigger(a.On), Key: a.Key, Text: a.Text,
		If: func(g *e.Game, c *e.Card, ev *e.Event) bool { return when(g, c, ev, whens) },
		Do: func(g *e.Game, c *e.Card, ev *e.Event) { run(g, c, ev, nil, effs) },
	})
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
	var gives e.Resource // FromDiscard decides what is generated
	if !a.FromDiscard {
		gives = resource(a.Gives)
	}
	sc.ResourceAbility = &e.ResourceAbility{
		Gives:       gives,
		FromDiscard: a.FromDiscard,
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

// applyStat adds a scripted stat bonus. Several stat abilities on one card
// add up; Max caps the stat including the card's printed value.
func applyStat(sc *e.Script, d *CardDoc, a Ability) {
	stat, bonus, prev := a.Stat, a.Bonus, sc.StatBonus
	limit := -1
	if a.Max > 0 {
		limit = a.Max - printedStat(d, stat)
	}
	sc.StatBonus = func(g *e.Game, c *e.Card, s string) int {
		n := 0
		if prev != nil {
			n = prev(g, c, s)
		}
		if s != stat {
			return n
		}
		b := evalCount(g, c, bonus)
		if limit >= 0 {
			b = min(b, limit)
		}
		return n + b
	}
}

// printedStat is the printed value a stat ability's Max applies to.
func printedStat(d *CardDoc, stat string) int {
	if stat == constants.StatHandSize {
		return d.HandSize
	}
	panic("cards: a max is only supported for " + constants.StatHandSize)
}

func actionOptions(abilities []Ability) func(*e.Game, *e.Card) []e.Option {
	return func(g *e.Game, c *e.Card) []e.Option {
		var opts []e.Option
		for _, a := range abilities {
			once := a.Key + constants.SepLabel + strconv.Itoa(c.ID)
			if !when(g, c, nil, a.When) || (a.OncePerRound && g.S.OncePerRound[once]) {
				continue
			}
			a := a
			do := func() {
				if a.OncePerRound {
					g.S.OncePerRound[once] = true
				}
				run(g, c, nil, nil, a.Effects)
			}
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
		sc.OnReveal != nil || sc.OnSetup != nil || sc.OnDefeated != nil || sc.BoostOnDamaged != nil ||
		sc.OnBoost != nil || sc.PlayWindow != "" || len(sc.Windows) > 0 ||
		sc.OnPlayEvent != nil || len(sc.Forced) > 0 || sc.Actions != nil ||
		sc.ResourceAbility != nil || sc.StatBonus != nil
}

var cardTypes = map[string]e.CardType{
	constants.TypeHero: e.TypeHero, constants.TypeAlterEgo: e.TypeAlterEgo, constants.TypeAlly: e.TypeAlly,
	constants.TypeEvent: e.TypeEvent, constants.TypeSupport: e.TypeSupport, constants.TypeUpgrade: e.TypeUpgrade,
	constants.TypeResource: e.TypeResource, constants.TypeVillain: e.TypeVillain, constants.TypeMainScheme: e.TypeMainScheme,
	constants.TypeSideScheme: e.TypeSideScheme, constants.TypeMinion: e.TypeMinion, constants.TypeTreachery: e.TypeTreachery,
	constants.TypeAttachment: e.TypeAttachment, constants.TypeObligation: e.TypeObligation,
	constants.TypeEnvironment: e.TypeEnvironment,
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
	constants.TrigAttacked:              e.TrigAttacked,
	constants.TrigAttackDamaged:         e.TrigAttackDamaged,
	constants.TrigMinionEnters:          e.TrigMinionEnters,
	constants.TrigEntersPlay:            e.TrigEntersPlay,
	constants.TrigEnemyDamaged:          e.TrigEnemyDamaged,
	constants.TrigVillainSchemed:        e.TrigVillainSchemed,
	constants.TrigStepOneThreat:         e.TrigStepOneThreat,
	constants.TrigHeroWouldTakeDamage:   e.TrigHeroWouldTakeDamage,
}

package cards

import (
	"fmt"
	"slices"

	"mcbench/constants"
	e "mcbench/game/engine"
)

// playerCardVerbs are effects of the player cards: heroes, allies and their
// events, supports and upgrades.
func playerCardVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbAttackMill:
		n := eff.Damage + eff.Amount*icons(mill(g, c, eff.N), resource(eff.Resource))
		attackEffect(g, c, ev, target, eff, n)
	case constants.VerbDamageAlly:
		g.DamageAlly(effectTarget(g, c, ev, target, eff), eff.Amount)
	case constants.VerbDamageAllEnemies:
		for _, en := range g.AllEnemies() {
			g.DamageEnemy(en, eff.Amount)
		}
	case constants.VerbDamageAllChars:
		damageFriendly(g, eff.Amount)
		for _, en := range g.AllEnemies() {
			g.DamageEnemy(en, eff.Amount)
		}
	case constants.VerbTempBoost:
		targets := []*e.Card{effectTarget(g, c, ev, target, eff)}
		if eff.Target == constants.SelFriendly {
			targets = friendly(g)
		}
		for _, t := range targets {
			tempBoost(g, t, eff.Stat, eff.Amount)
		}
	case constants.VerbGainTrait:
		g.S.PhaseTraits = append(g.S.PhaseTraits, eff.Trait)
		g.Logf("%s gains the %s trait until the end of the phase.", g.S.Hero.Name(), eff.Trait)
	case constants.VerbLookKeep:
		lookKeep(g, c, eff.N)
	case constants.VerbReturnFromDiscard:
		returnFromDiscard(g, eff.Trait)
	case constants.VerbDiscardHandChoice:
		discardHandChoice(g, c, ev, target, eff)
	case constants.VerbPutIntoPlay:
		g.PutIntoPlay(target)
	case constants.VerbMill:
		mill(g, c, eff.N)
	default:
		return false
	}
	return true
}

// encounterVerbs are effects of the encounter cards.
func encounterVerbs(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) bool {
	switch eff.Verb {
	case constants.VerbDamageHeroMill:
		g.HarmHero(eff.Amount*icons(mill(g, c, eff.N), resource(eff.Resource)), nil)
	case constants.VerbDamageFriendly:
		damageFriendly(g, eff.Amount)
	case constants.VerbExhaustUpgrades:
		exhaustAll(g, controlledUpgrades(g))
	case constants.VerbExhaustAllies:
		exhaustAll(g, allies(g))
	case constants.VerbExhaustCharacters:
		exhaustAll(g, friendly(g))
	case constants.VerbDrone:
		for i := 0; i < eff.N; i++ {
			g.PutDrone(droneFace())
		}
	case constants.VerbDiscardHandRes:
		discardHandResource(g, c, ev, target, eff)
	case constants.VerbDiscardUntilMinion:
		discardUntilMinion(g, eff.Trait)
	case constants.VerbSearchMinion:
		searchMinion(g, c, eff.Trait)
	case constants.VerbEngageSelf:
		g.S.EncDiscard = removeCard(g.S.EncDiscard, c)
		g.EngageMinion(c)
	case constants.VerbMinionsAttack:
		minionsAttack(g, c, ev, target, eff)
	default:
		return false
	}
	return true
}

// mill discards the top n cards of your deck. Like millKeep, it stops when
// the deck empties.
func mill(g *e.Game, c *e.Card, n int) []*e.Card {
	var out []*e.Card
	g.LastMilled = nil
	for i := 0; i < n && len(g.S.Deck) > 0; i++ {
		top := g.S.Deck[0]
		g.S.Deck = g.S.Deck[1:]
		g.S.Discard = append(g.S.Discard, top)
		g.LastMilled = top
		out = append(out, top)
		g.Logf("%s discards %s from the top of your deck.", c.Def.Name, top.Def.Name)
	}
	g.ResetEmptyDeck()
	return out
}

// icons counts the printed res resources on the cards.
func icons(cards []*e.Card, res e.Resource) (n int) {
	for _, x := range cards {
		for _, r := range x.Def.Resources {
			if r == res {
				n++
			}
		}
	}
	return
}

func damageFriendly(g *e.Game, n int) {
	for _, a := range allies(g) {
		g.DamageAlly(a, n)
	}
	g.HarmHero(n, nil)
}

func exhaustAll(g *e.Game, cards []*e.Card) {
	for _, x := range cards {
		x.Exhausted = true
	}
	if len(cards) > 0 {
		g.Logf("%d card(s) you control are exhausted.", len(cards))
	}
}

func tempBoost(g *e.Game, t *e.Card, stat string, n int) {
	switch stat {
	case constants.StatATK:
		t.TempATK += n
	case constants.StatTHW:
		t.TempTHW += n
	default:
		panic("cards: unknown temp_boost stat " + stat)
	}
	g.Logf("%s gets +%d %s until the end of the phase.", g.Label(t), n, stat)
}

// droneFace is the face of a facedown Drone minion.
func droneFace() *e.CardDef {
	d, err := Get(constants.CodeDrone)
	if err != nil {
		panic(err)
	}
	return d
}

// lookKeep looks at the top n cards of your deck, adds the one the player
// picks to your hand and discards the others.
func lookKeep(g *e.Game, c *e.Card, n int) {
	g.Do(func() {
		top := g.S.Deck[:min(n, len(g.S.Deck))]
		if len(top) == 0 {
			return
		}
		g.S.Deck = g.S.Deck[len(top):]
		var opts []e.Option
		seen := map[string]bool{}
		for _, x := range top {
			card := x
			if seen[card.Def.Name] {
				continue
			}
			seen[card.Def.Name] = true
			opts = append(opts, e.NewAction(constants.PrefixEffect+c.Name()+constants.SepTarget+card.Def.Name, "Add "+card.Def.Name+" to your hand", func() {
				g.S.Hand = append(g.S.Hand, card)
				g.Logf("%s adds %s to your hand.", c.Name(), card.Def.Name)
				for _, o := range top {
					if o != card {
						g.S.Discard = append(g.S.Discard, o)
						g.Logf("%s discards %s.", c.Name(), o.Def.Name)
					}
				}
				g.ResetEmptyDeck()
			}))
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: you look at the top %d card(s) of your deck. Choose one to add to your hand; the others are discarded.", c.Name(), len(top)), Options: opts})
	})
}

// returnFromDiscard returns the topmost upgrade with the trait in your
// discard pile to your hand.
func returnFromDiscard(g *e.Game, trait string) {
	for i := len(g.S.Discard) - 1; i >= 0; i-- {
		x := g.S.Discard[i]
		if x.Def.Type == e.TypeUpgrade && x.Def.HasTrait(trait) {
			g.S.Discard = removeCard(g.S.Discard, x)
			g.S.Hand = append(g.S.Hand, x)
			g.Logf("%s returns from your discard pile to your hand.", x.Def.Name)
			return
		}
	}
	g.Logf("There is no %s upgrade in your discard pile.", trait)
}

// discardHandChoice makes the player discard a card of their choice from
// hand, then runs Then.
func discardHandChoice(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	g.Do(func() {
		var opts []e.Option
		seen := map[string]bool{}
		for _, x := range g.S.Hand {
			card := x
			if seen[card.Def.Name] {
				continue
			}
			seen[card.Def.Name] = true
			opts = append(opts, e.NewAction(constants.PrefixEffect+c.Def.Name+constants.SepTarget+card.Def.Name, "Discard "+card.Def.Name, func() {
				g.DiscardFromHand(card)
				run(g, c, ev, target, eff.Then)
			}))
		}
		if len(opts) == 0 {
			return
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: c.Name() + ": choose a card in your hand to discard.", Options: opts})
	})
}

// discardHandResource discards each card in hand with a printed resource of
// the effect's type; with none, IfZero runs.
func discardHandResource(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	res := resource(eff.Resource)
	var hit []*e.Card
	for _, x := range g.S.Hand {
		if slices.Contains(x.Def.Resources, res) {
			hit = append(hit, x)
		}
	}
	for _, x := range hit {
		g.DiscardFromHand(x)
	}
	if len(hit) == 0 {
		run(g, c, ev, target, eff.IfZero)
	}
}

// discardUntilMinion discards encounter cards until a minion (with the
// trait, if set) is discarded, and puts it into play engaged with you. It
// gives up after going through every encounter card once.
func discardUntilMinion(g *e.Game, trait string) {
	for range len(g.S.EncDeck) + len(g.S.EncDiscard) {
		x := g.DrawEncounter()
		if x == nil {
			return
		}
		if x.Def.Type == e.TypeMinion && (trait == "" || x.Def.HasTrait(trait)) {
			g.Logf("%s is discarded from the encounter deck and enters play.", x.Def.Name)
			g.EngageMinion(x)
			return
		}
		g.Logf("%s is discarded from the encounter deck.", x.Def.Name)
		g.S.EncDiscard = append(g.S.EncDiscard, x)
	}
}

// searchMinion lets the player pick a minion with the trait from the
// encounter deck or discard pile; it enters play engaged with you and the
// encounter deck is shuffled.
func searchMinion(g *e.Game, c *e.Card, trait string) {
	g.Do(func() {
		var opts []e.Option
		seen := map[string]bool{}
		for _, zone := range []*[]*e.Card{&g.S.EncDeck, &g.S.EncDiscard} {
			for _, x := range *zone {
				if x.Def.Type != e.TypeMinion || !x.Def.HasTrait(trait) || seen[x.Def.Name] {
					continue
				}
				seen[x.Def.Name] = true
				card, zone := x, zone
				opts = append(opts, e.NewAction(constants.PrefixEffect+c.Def.Name+constants.SepTarget+card.Def.Name, card.Def.Name+" enters play engaged with you", func() {
					*zone = removeCard(*zone, card)
					g.Shuffle(g.S.EncDeck)
					g.EngageMinion(card)
				}))
			}
		}
		if len(opts) == 0 {
			g.Shuffle(g.S.EncDeck)
			return
		}
		g.Ask(&e.Decision{Kind: constants.KindChoice, Prompt: fmt.Sprintf("%s: choose a %s minion from the encounter deck or discard pile.", c.Def.Name, trait), Options: opts})
	})
}

// minionsAttack makes each engaged minion with the trait attack you. If none
// can attack (a stunned minion only discards its stun), IfZero runs.
func minionsAttack(g *e.Game, c *e.Card, ev *e.Event, target *e.Card, eff *Effect) {
	attackers := minionsWithTrait(g, eff.Trait)
	steps := []func(){}
	made := 0
	for _, m := range attackers {
		if !m.Stunned {
			made++
		}
		steps = append(steps, func() {
			if slices.Contains(g.S.Minions, m) {
				g.EnemyAttack(m, false, nil)
			}
		})
	}
	g.Do(steps...)
	if made == 0 {
		run(g, c, ev, target, eff.IfZero)
	}
}

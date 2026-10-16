package benchmark

import (
	"fmt"
	"io"
	"strings"

	"mcbench/constants"
	"mcbench/ui"
)

// Analyze writes, per scenario, how each configuration played: why its games
// were lost, how it used its turns and how it ended them. Scores say how well
// a player did; these behaviours suggest why (see docs/research.md). Records
// written before choices carried the board and the offered options show "-"
// for the turn measures.
func Analyze(w io.Writer, recs []Record, t ui.Theme) {
	groups := groupRecords(recs)
	keys := sortedKeys(groups)
	for i := 0; i < len(keys); {
		j := i
		for j < len(keys) && keys[j].scenario == keys[i].scenario {
			j++
		}
		var play, ends [][]string
		for _, k := range keys[i:j] {
			p, e := behaviourRows(k, groups[k])
			play, ends = append(play, p), append(ends, e)
		}
		lines := append(t.Table(playHeader, play), "")
		lines = append(lines, t.Table(endsHeader, ends)...)
		ui.Print(w, t.Box(keys[i].scenario+" · behaviour", lines, analyzeWidth))
		i = j
	}
	fmt.Fprintln(w, behaviourLegend)
}

// analyzeWidth fits the longest player and instruction names next to the measures.
const analyzeWidth = 150

// riskyThreatMargin: an alter-ego turn that ends with the main scheme this
// close to its target is risky, since the villain then schemes with a boost
// (in recorded games such a villain phase added 3-4 threat, up to 6).
const riskyThreatMargin = 4

var (
	playHeader = ui.Cols("player", "instructions", "graded", "won", "scheme", "defeated", "basic/rnd", "attack", "hero def")
	endsHeader = ui.Cols("player", "instructions", "turns", "AE ends", "AE risky", "idle hero", "idle ally", "card left", "acts/turn")
)

var behaviourLegend = fmt.Sprintf(`Lost to: scheme (main scheme completed), defeated (hero at 0 HP). basic/rnd: hero basic attacks and thwarts per round; attack: their share that attacked.
hero def: share of defences made by the hero (it then skips its next basic power). Share of turns that ended: in alter-ego (AE ends), in alter-ego within %d threat
of the target (AE risky), with a basic power (idle hero) or an ally attack or thwart (idle ally) unused, with a card that could still be played (card left).
acts/turn: actions before ending the turn.`, riskyThreatMargin)

// behaviour accumulates the measures of one configuration.
type behaviour struct {
	graded, won, scheme, defeated int
	rounds                        int
	basic, attacks                int
	defends, heroDefends          int
	ends, aeEnds, aeRisky         int
	idleHero, idleAlly, cardLeft  int
	actions                       int
	recorded                      bool // the choices carry the board and offered options
}

// behaviourRows are one configuration's rows of the two tables.
func behaviourRows(k groupKey, rs []Record) (play, ends []string) {
	var b behaviour
	for _, r := range rs {
		if r.Status == constants.InfraError {
			continue
		}
		b.add(r)
	}
	name := []string{k.player, instrName(k.instr)}
	play = append(name, fmt.Sprint(b.graded), fmt.Sprint(b.won), fmt.Sprint(b.scheme), fmt.Sprint(b.defeated),
		ratio2(b.basic, b.rounds), percent(b.attacks, b.basic), percent(b.heroDefends, b.defends))
	ends = append(name[:2:2], fmt.Sprint(b.ends))
	if !b.recorded {
		return play, append(ends, noValue, noValue, noValue, noValue, noValue, noValue)
	}
	return play, append(ends, percent(b.aeEnds, b.ends), percent(b.aeRisky, b.ends), percent(b.idleHero, b.ends),
		percent(b.idleAlly, b.ends), percent(b.cardLeft, b.ends), ratio2(b.actions, b.ends))
}

func (b *behaviour) add(r Record) {
	b.graded++
	b.rounds += r.Rounds
	switch {
	case r.Status == constants.Won:
		b.won++
	case r.Status == constants.Lost && r.Stats[constants.MetricHeroHP] <= 0:
		b.defeated++
	case r.Status == constants.Lost:
		b.scheme++
	}
	identity := identityNames(r)
	target := int(r.Stats[constants.MetricMainTarget])
	for _, c := range r.Choices {
		b.recorded = b.recorded || len(c.Offered) > 0
		switch {
		case c.Kind == constants.KindDefend:
			b.defends++
			if identity[strings.TrimPrefix(c.Key, constants.PrefixDefend)] {
				b.heroDefends++
			}
		case c.Key == constants.KeyEndTurn:
			b.endTurn(c.Form, c.Threat, target, c.Offered, identity)
		case c.Kind == constants.KindTurn:
			b.actions++
			if actor, ok := strikeActor(c.Key); ok && identity[actor] {
				b.basic++
				if strings.HasPrefix(c.Key, constants.PrefixAttack) {
					b.attacks++
				}
			}
		}
	}
}

// endTurn classifies how a turn ended from the board and the options left.
func (b *behaviour) endTurn(form string, threat, target int, offered []string, identity map[string]bool) {
	b.ends++
	if form == constants.TypeAlterEgo {
		b.aeEnds++
		if threat+riskyThreatMargin >= target {
			b.aeRisky++
		}
	}
	var hero, ally, card bool
	for _, k := range offered {
		actor, strike := strikeActor(k)
		hero = hero || (strike && identity[actor]) || strings.HasPrefix(k, constants.PrefixRecover)
		ally = ally || (strike && !identity[actor])
		card = card || strings.HasPrefix(k, constants.PrefixPlay)
	}
	b.idleHero += count(hero)
	b.idleAlly += count(ally)
	b.cardLeft += count(card)
}

// strikeActor returns who attacks or thwarts in an attack:/thwart: key.
func strikeActor(key string) (string, bool) {
	for _, p := range []string{constants.PrefixAttack, constants.PrefixThwart} {
		if rest, ok := strings.CutPrefix(key, p); ok {
			actor, _, _ := strings.Cut(rest, constants.SepTarget)
			return actor, true
		}
	}
	return "", false
}

// identityNames are the names of the player's identity (both forms), read
// from the change-form options: every game offers them.
func identityNames(r Record) map[string]bool {
	names := map[string]bool{}
	for _, c := range r.Choices {
		for _, k := range append([]string{c.Key}, c.Offered...) {
			if name, ok := strings.CutPrefix(k, constants.PrefixChangeForm); ok {
				names[name] = true
			}
		}
	}
	return names
}

func count(b bool) int {
	if b {
		return 1
	}
	return 0
}

func percent(n, d int) string {
	if d == 0 {
		return noValue
	}
	return fmt.Sprintf("%.0f%%", 100*float64(n)/float64(d))
}

func ratio2(n, d int) string {
	if d == 0 {
		return noValue
	}
	return fmt.Sprintf("%.2f", float64(n)/float64(d))
}

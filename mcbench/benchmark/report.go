package benchmark

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"mcbench/constants"
	"mcbench/ui"
)

// groupKey identifies one configuration (player + instruction set).
type groupKey struct{ player, instr, scenario string }

// criterionLabels are the short column headers for CriteriaOrder.
var criterionLabels = map[string]string{
	constants.CriterionWin:           "win",
	constants.CriterionVillainDamage: "dmg",
	constants.CriterionHeroHP:        "hp",
	constants.CriterionThreat:        "threat",
	constants.CriterionSpeed:         "speed",
}

const (
	meterCells = 10 // cells in a score meter
	noValue    = "-"
)

// Report writes, per scenario, one panel per configuration with its aggregate
// and per-game results, then the overview of every configuration last so it
// stays visible at the bottom of the terminal.
func Report(w io.Writer, recs []Record, t ui.Theme) {
	groups := groupRecords(recs)
	keys := sortedKeys(groups)
	var overviews [][]string
	for i := 0; i < len(keys); {
		j := i
		for j < len(keys) && keys[j].scenario == keys[i].scenario {
			j++
		}
		for _, k := range keys[i:j] {
			ui.Print(w, configPanel(k, groups[k], t))
		}
		overviews = append(overviews, overviewPanel(keys[i:j], groups, t))
		i = j
	}
	writeTotals(w, recs)
	ui.Print(w, overviews...)
}

func groupRecords(recs []Record) map[groupKey][]Record {
	groups := map[groupKey][]Record{}
	for _, r := range recs {
		groups[keyOf(r)] = append(groups[keyOf(r)], r)
	}
	return groups
}

func keyOf(r Record) groupKey {
	instr := strings.Join(r.Instructions.Names, ",")
	if instr == "" {
		instr = "(none)"
	} else {
		instr += "@" + r.Instructions.SHA256
	}
	return groupKey{r.Player, instr, r.Scenario}
}

func sortedKeys(groups map[groupKey][]Record) []groupKey {
	keys := make([]groupKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.scenario != b.scenario {
			return a.scenario < b.scenario
		}
		if a.player != b.player {
			return a.player < b.player
		}
		return a.instr < b.instr
	})
	return keys
}

// overviewPanel compares the configurations of one scenario: mean score and
// mean criteria (win = win rate).
func overviewPanel(keys []groupKey, groups map[groupKey][]Record, t ui.Theme) []string {
	cols := ui.Cols("player", "instructions", "graded", "score ± s.e.")
	for _, c := range CriteriaOrder {
		cols = append(cols, ui.Column{Title: criterionLabels[c]})
	}
	var rows [][]string
	for _, k := range keys {
		s := summarize(groups[k])
		row := []string{k.player, instrName(k.instr), fmt.Sprint(len(s.score.xs)),
			fmt.Sprintf("%.3f ± %.3f", s.score.mean(), s.score.stderr())}
		for _, c := range CriteriaOrder {
			row = append(row, fmt.Sprintf("%.2f", s.criteria[c].mean()))
		}
		rows = append(rows, row)
	}
	return t.Box(keys[0].scenario, t.Table(cols, rows), 0)
}

// configPanel shows one configuration: how its games ended, what they cost and
// a row per game.
func configPanel(k groupKey, rs []Record, t ui.Theme) []string {
	s := summarize(rs)
	games := sortedGames(rs)
	lines := []string{
		t.Dim("instructions " + k.instr),
		fmt.Sprintf("%d games: %s", len(rs), statusText(s.status)),
		fmt.Sprintf("score %.3f ± %.3f  %s", s.score.mean(), s.score.stderr(), ui.Meter(s.score.mean(), meterCells)),
		aggregateLine(s, rs),
		"",
	}
	lines = append(lines, t.Table(gameHeader, gameRows(games, t))...)
	for _, r := range games {
		if r.Error != "" {
			lines = append(lines, t.Red(fmt.Sprintf("seed %s: %s", gameLabel(r), firstLine(r.Error))))
		}
	}
	return t.Box(k.player+" · "+k.scenario, lines, ui.Width)
}

// aggregateLine summarizes length and cost per game.
func aggregateLine(s summary, rs []Record) string {
	won := noValue
	if len(s.roundsWon.xs) > 0 {
		won = fmt.Sprintf("%.1f", s.roundsWon.mean())
	}
	var invalid int
	var in, out, ms int64
	for _, r := range rs {
		invalid += r.Invalid
		in += r.Usage.InputTokens
		out += r.Usage.OutputTokens
		ms += r.Millis
	}
	n := int64(len(rs))
	return fmt.Sprintf("rounds to win %s · decisions %.0f · invalid %d · per game: tokens %s in / %s out, %s",
		won, s.decisions.mean(), invalid, tokens(in/n), tokens(out/n), elapsed(ms/n))
}

// gameHeader labels gameRows: result, score, the final board facts the
// criteria are computed from (villain damage, hero hp, main-scheme threat) and
// cost (invalid choices, input tokens, wall time).
var gameHeader = append(ui.Cols("seed", "result", "score", "", "rounds", "dec",
	"villain", "hero", "threat", "inv"), ui.Column{Title: "tokens", Right: true}, ui.Column{Title: "time"})

func gameRows(games []Record, t ui.Theme) [][]string {
	var rows [][]string
	for _, r := range games {
		rows = append(rows, []string{
			gameLabel(r), statusColor(r.Status, t), fmt.Sprintf("%.3f", r.Score), ui.Meter(r.Score, meterCells),
			fmt.Sprint(r.Rounds), fmt.Sprint(r.Decisions),
			fraction(r.Stats, constants.MetricVillainDamage, constants.MetricVillainTotalHP),
			fraction(r.Stats, constants.MetricHeroHP, constants.MetricHeroMaxHP),
			fraction(r.Stats, constants.MetricMainThreat, constants.MetricMainTarget),
			fmt.Sprint(r.Invalid), tokens(r.Usage.InputTokens), elapsed(r.Millis),
		})
	}
	return rows
}

func sortedGames(rs []Record) []Record {
	out := append([]Record(nil), rs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Seed != out[j].Seed {
			return out[i].Seed < out[j].Seed
		}
		return out[i].Sample < out[j].Sample
	})
	return out
}

func gameLabel(r Record) string {
	if r.Sample == 0 {
		return fmt.Sprint(r.Seed)
	}
	return fmt.Sprintf("%d/s%d", r.Seed, r.Sample)
}

func statusColor(status string, t ui.Theme) string {
	switch status {
	case constants.Won:
		return t.Green(status)
	case constants.Lost:
		return t.Red(status)
	}
	return t.Yellow(status)
}

// fraction renders stats[num]/stats[den], e.g. hero hp "6/10".
func fraction(stats map[string]float64, num, den string) string {
	return fmt.Sprintf("%.0f/%.0f", stats[num], stats[den])
}

// instrName drops the hash from an instruction key; the config panel shows it.
func instrName(instr string) string {
	name, _, _ := strings.Cut(instr, "@")
	return name
}

// tokens renders a token count; scripted players use none.
func tokens(n int64) string {
	if n == 0 {
		return noValue
	}
	return ui.Thousands(n)
}

func elapsed(ms int64) string { return ui.Elapsed(time.Duration(ms) * time.Millisecond) }

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

type summary struct {
	score     stat
	criteria  map[string]*stat
	roundsWon stat
	decisions stat
	status    map[string]int
}

// summarize aggregates the graded games and counts every status. A game cut
// short by an infrastructure failure is counted but not graded.
func summarize(rs []Record) summary {
	s := summary{criteria: map[string]*stat{}, status: map[string]int{}}
	for _, c := range CriteriaOrder {
		s.criteria[c] = &stat{}
	}
	for _, r := range rs {
		s.status[r.Status]++
		if r.Status == constants.InfraError {
			continue
		}
		s.score.add(r.Score)
		for _, c := range CriteriaOrder {
			s.criteria[c].add(r.Criteria[c])
		}
		if r.Status == constants.Won {
			s.roundsWon.add(float64(r.Rounds))
		}
		s.decisions.add(float64(r.Decisions))
	}
	return s
}

func statusText(status map[string]int) string {
	var parts []string
	for s, n := range status {
		parts = append(parts, fmt.Sprintf("%s %d", s, n))
	}
	sort.Strings(parts)
	return strings.Join(parts, " · ")
}

func writeTotals(w io.Writer, recs []Record) {
	var invalid int
	var in, out int64
	for _, r := range recs {
		invalid += r.Invalid
		in += r.Usage.InputTokens
		out += r.Usage.OutputTokens
	}
	fmt.Fprintf(w, "Criteria are means over games (win = win rate). Invalid choices: %d. Tokens: %s in, %s out.\n", invalid, ui.Thousands(in), ui.Thousands(out))
}

// stat accumulates numbers and reports their mean and standard error.
type stat struct{ xs []float64 }

func (s *stat) add(x float64) { s.xs = append(s.xs, x) }

func (s stat) mean() float64 {
	if len(s.xs) == 0 {
		return math.NaN()
	}
	total := 0.0
	for _, x := range s.xs {
		total += x
	}
	return total / float64(len(s.xs))
}

func (s stat) stderr() float64 {
	n := float64(len(s.xs))
	if n < 2 {
		return 0
	}
	m := s.mean()
	v := 0.0
	for _, x := range s.xs {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v/(n-1)) / math.Sqrt(n)
}

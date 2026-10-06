package benchmark

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"mcbench/constants"
)

// groupKey identifies one configuration (player + instruction set).
type groupKey struct{ player, instr, scenario string }

// Report writes a Markdown summary per configuration and scenario.
func Report(w io.Writer, recs []Record) {
	groups := groupRecords(recs)
	scenario := ""
	for _, k := range sortedKeys(groups) {
		if k.scenario != scenario {
			scenario = k.scenario
			fmt.Fprintf(w, "\n## %s\n\n%s\n", scenario, tableHeader())
		}
		fmt.Fprintln(w, groupRow(k, groups[k]))
	}
	writeTotals(w, recs)
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

func tableHeader() string {
	header := "| player | instructions | games | score | ± s.e. |"
	sep := "|---|---|---|---|---|"
	for _, c := range CriteriaOrder {
		header += " " + c + " |"
		sep += "---|"
	}
	return header + " rounds (wins) | decisions | status |\n" + sep + "---|---|---|"
}

type summary struct {
	score     stat
	criteria  map[string]*stat
	roundsWon stat
	decisions stat
	status    map[string]int
}

func summarize(rs []Record) summary {
	s := summary{criteria: map[string]*stat{}, status: map[string]int{}}
	for _, r := range rs {
		s.score.add(r.Score)
		for _, c := range CriteriaOrder {
			if s.criteria[c] == nil {
				s.criteria[c] = &stat{}
			}
			s.criteria[c].add(r.Criteria[c])
		}
		if r.Status == constants.Won {
			s.roundsWon.add(float64(r.Rounds))
		}
		s.decisions.add(float64(r.Decisions))
		s.status[r.Status]++
	}
	return s
}

func groupRow(k groupKey, rs []Record) string {
	s := summarize(rs)
	row := fmt.Sprintf("| %s | %s | %d | %.3f | %.3f |", k.player, k.instr, len(rs), s.score.mean(), s.score.stderr())
	for _, c := range CriteriaOrder {
		row += fmt.Sprintf(" %.2f |", s.criteria[c].mean())
	}
	won := "-"
	if len(s.roundsWon.xs) > 0 {
		won = fmt.Sprintf("%.1f", s.roundsWon.mean())
	}
	return row + fmt.Sprintf(" %s | %.0f | %s |", won, s.decisions.mean(), statusText(s.status))
}

func statusText(status map[string]int) string {
	var parts []string
	for s, n := range status {
		parts = append(parts, fmt.Sprintf("%s=%d", s, n))
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

func writeTotals(w io.Writer, recs []Record) {
	var invalid int
	var in, out int64
	for _, r := range recs {
		invalid += r.Invalid
		in += r.Usage.InputTokens
		out += r.Usage.OutputTokens
	}
	fmt.Fprintf(w, "\nCriteria are means over games (win = win rate). Invalid choices: %d. Tokens: %d in, %d out.\n", invalid, in, out)
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

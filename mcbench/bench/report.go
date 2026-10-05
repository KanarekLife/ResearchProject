package bench

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"mcbench/game"
)

// ReadRecords reads a games.jsonl file.
func ReadRecords(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<30)
	for sc.Scan() {
		var r Record
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, sc.Err()
}

type stat struct{ xs []float64 }

func (s *stat) add(x float64) { s.xs = append(s.xs, x) }

func (s stat) mean() float64 {
	if len(s.xs) == 0 {
		return math.NaN()
	}
	t := 0.0
	for _, x := range s.xs {
		t += x
	}
	return t / float64(len(s.xs))
}

// stderr is the standard error of the mean.
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

// Report writes a Markdown summary per configuration (agent + instruction
// set) and scenario.
func Report(w io.Writer, recs []Record) {
	type key struct{ agent, instr, scenario string }
	groups := map[key][]Record{}
	for _, r := range recs {
		instr := strings.Join(r.Instructions.Names, ",")
		if instr == "" {
			instr = "(none)"
		} else {
			instr += "@" + r.Instructions.SHA256
		}
		k := key{r.Agent, instr, r.Scenario}
		groups[k] = append(groups[k], r)
	}
	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.scenario != b.scenario {
			return a.scenario < b.scenario
		}
		if a.agent != b.agent {
			return a.agent < b.agent
		}
		return a.instr < b.instr
	})

	header := "| agent | instructions | games | score | ± s.e. |"
	sep := "|---|---|---|---|---|"
	for _, c := range CriteriaOrder {
		header += " " + c + " |"
		sep += "---|"
	}
	header += " rounds (wins) | decisions | status |"
	sep += "---|---|---|"

	scenario := ""
	for _, k := range keys {
		if k.scenario != scenario {
			scenario = k.scenario
			fmt.Fprintf(w, "\n## %s\n\n%s\n%s\n", scenario, header, sep)
		}
		rs := groups[k]
		var score stat
		crit := map[string]*stat{}
		var roundsWon, decisions stat
		status := map[string]int{}
		for _, r := range rs {
			score.add(r.Score)
			for _, c := range CriteriaOrder {
				if crit[c] == nil {
					crit[c] = &stat{}
				}
				crit[c].add(r.Criteria[c])
			}
			if r.Status == game.Won {
				roundsWon.add(float64(r.Rounds))
			}
			decisions.add(float64(r.Decisions))
			status[r.Status]++
		}
		line := fmt.Sprintf("| %s | %s | %d | %.3f | %.3f |", k.agent, k.instr, len(rs), score.mean(), score.stderr())
		for _, c := range CriteriaOrder {
			line += fmt.Sprintf(" %.2f |", crit[c].mean())
		}
		rw := "-"
		if len(roundsWon.xs) > 0 {
			rw = fmt.Sprintf("%.1f", roundsWon.mean())
		}
		var st []string
		for s, n := range status {
			st = append(st, fmt.Sprintf("%s=%d", s, n))
		}
		sort.Strings(st)
		line += fmt.Sprintf(" %s | %.0f | %s |", rw, decisions.mean(), strings.Join(st, " "))
		fmt.Fprintln(w, line)
	}

	var invalid int
	var tokIn, tokOut int64
	for _, r := range recs {
		invalid += r.Invalid
		tokIn += r.Usage.InputTokens
		tokOut += r.Usage.OutputTokens
	}
	fmt.Fprintf(w, "\nCriteria are means over games (win = win rate). Invalid choices: %d. Tokens: %d in, %d out.\n", invalid, tokIn, tokOut)
}

package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Run is one results directory: its id, where it lives and the games in it.
type Run struct {
	ID    string
	Dir   string
	Games []Game
}

// Game is one game's trace files inside a run.
type Game struct {
	Name  string // file base, e.g. spider-man-vs-rhino_seed2_s0
	Trace string
	Live  string
}

// Discover lists the runs under dir, newest first (run ids sort by timestamp).
func Discover(dir string) ([]Run, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var runs []Run
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rdir := filepath.Join(dir, e.Name())
		games, err := games(rdir)
		if err != nil {
			return nil, err
		}
		if len(games) == 0 {
			continue // a run with no trace yet is not worth showing
		}
		runs = append(runs, Run{ID: e.Name(), Dir: rdir, Games: games})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].ID > runs[j].ID })
	return runs, nil
}

// games lists a run's trace files, sorted by name.
func games(dir string) ([]Game, error) {
	live := filepath.Join(dir, liveSubdir)
	entries, err := os.ReadDir(live)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Game
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".jsonl")
		out = append(out, Game{
			Name:  base,
			Trace: filepath.Join(live, e.Name()),
			Live:  filepath.Join(live, base+".txt"),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// PickRun resolves a run selector: empty means the newest run, otherwise an
// exact id or a unique prefix.
func PickRun(runs []Run, sel string) (Run, error) {
	if len(runs) == 0 {
		return Run{}, fmt.Errorf("no runs under the results directory")
	}
	if sel == "" {
		return runs[0], nil
	}
	for _, r := range runs {
		if r.ID == sel {
			return r, nil
		}
	}
	var hits []Run
	for _, r := range runs {
		if strings.HasPrefix(r.ID, sel) {
			hits = append(hits, r)
		}
	}
	switch len(hits) {
	case 0:
		return Run{}, fmt.Errorf("no run matches %q", sel)
	case 1:
		return hits[0], nil
	default:
		return Run{}, fmt.Errorf("%q matches %d runs", sel, len(hits))
	}
}

// PickGame resolves a game selector against a run: empty means the only game, or
// the last one by name when there are several. Otherwise an exact name or a
// unique prefix.
func PickGame(r Run, sel string) (Game, error) {
	if len(r.Games) == 0 {
		return Game{}, fmt.Errorf("run %s has no traces", r.ID)
	}
	if sel == "" {
		return r.Games[len(r.Games)-1], nil
	}
	for _, g := range r.Games {
		if g.Name == sel {
			return g, nil
		}
	}
	var hits []Game
	for _, g := range r.Games {
		if strings.Contains(g.Name, sel) {
			hits = append(hits, g)
		}
	}
	switch len(hits) {
	case 0:
		return Game{}, fmt.Errorf("no game in %s matches %q", r.ID, sel)
	case 1:
		return hits[0], nil
	default:
		return Game{}, fmt.Errorf("%q matches %d games", sel, len(hits))
	}
}

// liveSubdir is where a run keeps its per-game trace files.
const liveSubdir = "live"

package boardgui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"mcbench/game/scenario"
	"mcbench/trace"
)

// Export writes the board as a static site in out, a new directory, with every
// API response precomputed, so it can be hosted without the server. A game the
// current engine cannot replay keeps its board from the fallback site, an
// export made with an earlier engine, when that one replayed it.
func Export(out, fallback string, d Data) error {
	page, err := fs.Sub(static, "static")
	if err != nil {
		return err
	}
	if _, err := fs.Stat(page, "index.html"); err != nil {
		return errors.New("the board page is not built; from mcbench/, run: go generate ./boardgui")
	}
	if err := os.CopyFS(out, page); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(out, notBuilt)); err != nil {
		return err
	}
	runs, err := trace.Discover(d.Results...)
	if err != nil {
		return err
	}
	earlier, err := readSite(fallback)
	if err != nil {
		return err
	}
	list := Runs{Static: true}
	if len(runs) > 0 {
		list.Run = runs[0].ID
	}
	for i, r := range runs {
		info := RunInfo{ID: r.ID, Version: version(r.Dir)}
		for j, g := range r.Games {
			// Files are numbered: run ids and game names need not be URL-safe.
			file := fmt.Sprintf("api/steps/%d-%d.json", i, j)
			snap, err := replayFile(g.Trace, d.Scenarios)
			if err != nil {
				return err
			}
			snap = orEarlier(snap, earlier[[2]string{r.ID, g.Name}])
			snap.Run, snap.Game = r.ID, g.Name
			if err := writeFile(filepath.Join(out, file), snap); err != nil {
				return err
			}
			state := stateLive
			if snap.End != nil {
				state = snap.End.Status
			}
			info.Games = append(info.Games, GameItem{Name: g.Name, State: state, Steps: file})
		}
		list.Runs = append(list.Runs, info)
	}
	stats, err := d.Stats()
	if err != nil {
		return err
	}
	// The pages ask the server for these, so they keep its routes.
	for name, v := range map[string]any{"runs": list, "cards": cardCodes(), "stats": stats} {
		if err := writeFile(filepath.Join(out, "api", name), v); err != nil {
			return err
		}
	}
	return nil
}

func replayFile(path string, scs []*scenario.Scenario) (Snapshot, error) {
	entries, err := trace.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}
	r := newReplay(scs)
	r.apply(entries)
	return r.snapshot(0), nil
}

// orEarlier is snap, or the earlier export's replay when only that one
// replayed the whole trace.
func orEarlier(snap, earlier Snapshot) Snapshot {
	if snap.Error != "" && earlier.Total > 0 && earlier.Error == "" {
		return earlier
	}
	return snap
}

// readSite reads every game's replay from an exported site, by run and game.
// An empty dir reads nothing.
func readSite(dir string) (map[[2]string]Snapshot, error) {
	snaps := map[[2]string]Snapshot{}
	if dir == "" {
		return snaps, nil
	}
	var list Runs
	if err := readFile(filepath.Join(dir, "api", "runs"), &list); err != nil {
		return nil, err
	}
	for _, r := range list.Runs {
		for _, g := range r.Games {
			var snap Snapshot
			if err := readFile(filepath.Join(dir, g.Steps), &snap); err != nil {
				return nil, err
			}
			snaps[[2]string{r.ID, g.Name}] = snap
		}
	}
	return snaps, nil
}

func readFile(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

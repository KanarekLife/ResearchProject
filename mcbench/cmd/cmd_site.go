package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"mcbench/benchmark"
	"mcbench/boardgui"
	"mcbench/data"
	"mcbench/game/scenario"
	"mcbench/trace"
)

const defaultSiteDir = "site"

// cmdSite writes the board and the statistics page as a static website.
func cmdSite(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameSite, args)
	if err != nil {
		return err
	}
	fs.Usage = func() {
		fs.Output().Write([]byte("usage: mcbench site [-out DIR] [RESULTS_DIR]...\n"))
		fs.PrintDefaults()
	}
	out := fs.String("out", defaultSiteDir, "new directory to write the site to")
	source := sourceFlag(fs, cfg)
	fs.Parse(args)
	dirs := fs.Args()
	if len(dirs) == 0 {
		if dirs, err = resultDirs(*source, cfg.Run.Out, store(cfg)); err != nil {
			return err
		}
	}
	scs, err := scenario.LoadAll(data.FS)
	if err != nil {
		return err
	}
	if err := boardgui.Export(*out, boardData(dirs, "", scs)); err != nil {
		return err
	}
	fmt.Printf("wrote the site to %s\n", *out)
	return nil
}

func boardData(dirs []string, online string, scs []*scenario.Scenario) boardgui.Data {
	return boardgui.Data{Results: dirs, Online: online, Scenarios: scs,
		Stats: func() (any, error) { return gameStats(dirs) }}
}

// gameStats names runs as the board does, so a game links to its board.
func gameStats(dirs []string) ([]benchmark.GameStats, error) {
	runs, err := trace.Discover(dirs...)
	if err != nil {
		return nil, err
	}
	out := []benchmark.GameStats{}
	for _, r := range runs {
		recs, err := benchmark.ReadRecords(filepath.Join(r.Dir, gamesFile))
		if errors.Is(err, fs.ErrNotExist) {
			continue // no game has finished yet
		}
		if err != nil {
			return nil, err
		}
		for _, rec := range recs {
			out = append(out, benchmark.StatsOf(r.ID, rec))
		}
	}
	return out, nil
}

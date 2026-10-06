package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"mcbench/boardgui"
	"mcbench/trace"
)

// defaultBoardAddr keeps the board on localhost.
const defaultBoardAddr = "127.0.0.1:8090"

func cmdBoardGUI(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameBoardGUI, args)
	if err != nil {
		return err
	}
	var (
		addr    = fs.String("addr", defaultBoardAddr, "listen address (keep it on localhost)")
		dir     = fs.String("results", cfg.Run.Out, "results directory, used when no path is given")
		runSel  = fs.String("run", "", "run id or prefix (default: newest)")
		gameSel = fs.String("game", "", "game name or substring (default: last)")
		ld      = scenarioFlags(fs, cfg)
	)
	fs.Parse(args)
	path := fs.Arg(0)
	fs.Parse(fs.Args()[min(1, fs.NArg()):]) // flags may also follow the path

	results, run, game, err := resolveSelection(path, *dir, *runSel, *gameSel)
	if err != nil {
		return err
	}
	scs, err := ld.load()
	if err != nil {
		return err
	}
	shown := game
	if shown == "" {
		shown = "its current game"
	}
	fmt.Printf("viewing run %s, %s, at http://%s/ · ctrl-c to stop\n", run, shown, *addr)
	return boardgui.Serve(*addr, results, run, game, scs)
}

// resolveSelection turns the command line into the results directory and the
// run and game the page opens first. A trace file or -game pins that game;
// otherwise game is empty and the page follows the run's current game.
func resolveSelection(path, results, runSel, gameSel string) (string, string, string, error) {
	if strings.HasSuffix(path, ".jsonl") {
		// A trace lives at <results>/<run>/live/<game>.jsonl.
		runDir := filepath.Dir(filepath.Dir(filepath.Clean(path)))
		results, runSel = filepath.Dir(runDir), filepath.Base(runDir)
		gameSel = strings.TrimSuffix(filepath.Base(path), ".jsonl")
	} else if path != "" {
		path = filepath.Clean(path)
		results, runSel = filepath.Dir(path), filepath.Base(path)
	}
	runs, err := trace.Discover(results)
	if err != nil {
		return "", "", "", err
	}
	run, err := trace.PickRun(runs, runSel)
	if err != nil {
		return "", "", "", err
	}
	if gameSel == "" {
		return results, run.ID, "", nil
	}
	g, err := trace.PickGame(run, gameSel)
	if err != nil {
		return "", "", "", err
	}
	if strings.HasSuffix(path, ".jsonl") && g.Trace != filepath.Clean(path) {
		return "", "", "", fmt.Errorf("%s is not a game trace under a run's live/ directory", path)
	}
	return results, run.ID, g.Name, nil
}

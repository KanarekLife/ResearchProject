package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mcbench/benchmark"
	"mcbench/trace"
)

func cmdView(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameView, args)
	if err != nil {
		return err
	}
	var (
		dir    = fs.String("results", cfg.Run.Out, "results directory")
		runSel = fs.String("run", "", "run id or prefix (default: newest)")
		game   = fs.String("game", "", "game name or substring (default: last)")
		list   = fs.Bool("list", false, "list runs and their games")
		final  = fs.Bool("final", false, "summarize a run's finished games")
		follow = fs.Bool("follow", false, "keep printing decisions as the game is played")
	)
	fs.Parse(args)

	runs, err := trace.Discover(*dir)
	if err != nil {
		return err
	}
	if *list {
		listRuns(runs)
		return nil
	}
	run, err := trace.PickRun(runs, *runSel)
	if err != nil {
		return err
	}
	if *final {
		return viewSummary(run)
	}
	g, err := trace.PickGame(run, *game)
	if err != nil {
		return err
	}
	if *follow {
		return followGame(g)
	}
	return printGame(g)
}

// listRuns prints every run and the games it holds.
func listRuns(runs []trace.Run) {
	if len(runs) == 0 {
		fmt.Println("no runs found")
		return
	}
	for _, r := range runs {
		fmt.Printf("%s  %d game(s)\n", r.ID, len(r.Games))
		for _, g := range r.Games {
			fmt.Printf("  %s\n", g.Name)
		}
	}
}

// printGame renders a game's whole trace.
func printGame(g trace.Game) error {
	entries, err := trace.ReadFile(g.Trace)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, e := range entries {
		trace.Write(&b, e)
	}
	fmt.Print(b.String())
	return nil
}

// followGame prints entries as they are appended, until the game ends. It reads
// the trace the same way the game writes it, so it can attach to a game already
// in progress and needs no restart once the game finishes.
func followGame(g trace.Game) error {
	f, err := trace.Follow(g.Trace)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("following %s · ctrl-c to stop\n\n", g.Name)
	for {
		entries, err := f.Next(viewPoll)
		if err != nil {
			return err
		}
		var b strings.Builder
		for _, e := range entries {
			trace.Write(&b, e)
			if trace.IsEnd(e) {
				fmt.Print(b.String())
				return nil
			}
		}
		if b.Len() > 0 {
			fmt.Print(b.String())
		}
	}
}

// viewSummary prints the finished-game table for a run.
func viewSummary(run trace.Run) error {
	recs, err := benchmark.ReadRecords(filepath.Join(run.Dir, gamesFile))
	if err != nil {
		return err
	}
	benchmark.Report(os.Stdout, recs)
	return nil
}

// viewPoll is how long a follower waits for a trace to grow before ending one
// poll cycle.
const viewPoll = 250 * time.Millisecond

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
)

// cmdValidate plays every seed with the scripted players. Careless play
// reaches many engine paths, so any engine error here is a bug.
func cmdValidate(args []string) error {
	cfg, err := configFromArgs(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(cmdNameValidate, flag.ExitOnError)
	l := scenarioFlags(fs, cfg)
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	players := []player.Player{heuristic.Random{Seed: 1}, heuristic.First{}, heuristic.Heuristic{}}
	failed := false
	for _, p := range players {
		var recs []benchmark.Record
		benchmark.Run(context.Background(), scs, p, instruction.Instructions{}, benchmark.Config{Parallel: 4}, benchmark.DefaultWeights, func(r benchmark.Record) {
			recs = append(recs, r)
			if r.Status == session.EngineError || r.Status == constants.AgentError || r.Status == session.DecisionLimit {
				failed = true
				fmt.Printf("%s seed %d (%s): %s\n  %s\n", r.Scenario, r.Seed, r.Player, r.Status, firstLines(r.Error, 8))
			}
		})
		benchmark.Report(os.Stdout, recs)
	}
	if failed {
		return fmt.Errorf("engine errors found")
	}
	return nil
}

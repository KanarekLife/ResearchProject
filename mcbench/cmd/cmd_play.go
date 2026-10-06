package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"mcbench/benchmark"
	"mcbench/player/human"
	"mcbench/player/instruction"
)

func cmdPlay(args []string) error {
	fs, cfg, err := newFlagSet(cmdNamePlay, args)
	if err != nil {
		return err
	}
	l := scenarioFlags(fs, cfg)
	seed := fs.Uint64("seed", 0, "seed (default: the scenario's first)")
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	if *seed == 0 {
		*seed = scs[0].Seeds[0]
	}
	p := &human.Human{In: bufio.NewReader(os.Stdin), Out: os.Stdout}
	rec := benchmark.Play(context.Background(), scs[0], p, instruction.Instructions{}, *seed, 0, "", benchmark.DefaultWeights)
	out, _ := json.MarshalIndent(map[string]any{"status": rec.Status, "rounds": rec.Rounds, "criteria": rec.Criteria, "score": rec.Score}, "", "  ")
	fmt.Println(string(out))
	return nil
}

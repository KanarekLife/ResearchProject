package main

import (
	"flag"
	"fmt"

	"mcbench/constants"
)

func cmdShow(args []string) error {
	cfg, err := configFromArgs(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(cmdNameShow, flag.ExitOnError)
	l := scenarioFlags(fs, cfg)
	seed := fs.Uint64("seed", 0, "seed (default: the scenario's first)")
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	s, err := newSession(scs[0], *seed)
	if err != nil {
		return err
	}
	fmt.Printf("%s:\n%s\n\n%s:\n%s\n",
		constants.ToolGetState, s.Call(constants.ToolGetState, nil),
		constants.ToolGetDecision, s.Call(constants.ToolGetDecision, nil))
	return nil
}

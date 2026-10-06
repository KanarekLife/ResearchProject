package main

import (
	"fmt"
)

func cmdList(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameList, args)
	if err != nil {
		return err
	}
	l := scenarioFlags(fs, cfg)
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	for _, s := range scs {
		fmt.Printf("%s: %s (%d seeds, max %d rounds)\n", s.ID, s.Title, len(s.Seeds), s.MaxRounds)
	}
	return nil
}

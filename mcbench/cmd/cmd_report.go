package main

import (
	"fmt"
	"os"

	"mcbench/benchmark"
)

func cmdReport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcbench report GAMES.jsonl...")
	}
	var all []benchmark.Record
	for _, p := range args {
		recs, err := benchmark.ReadRecords(p)
		if err != nil {
			return err
		}
		all = append(all, recs...)
	}
	benchmark.Report(os.Stdout, all)
	return nil
}

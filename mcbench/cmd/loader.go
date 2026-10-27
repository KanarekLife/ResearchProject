package main

import (
	"mcbench/constants"

	"flag"
	"fmt"
	"strings"

	"mcbench/data"
	"mcbench/game/scenario"
)

// loader holds the scenario selection flag shared by most commands.
type loader struct{ only *string }

// newFlagSet loads the config file named by -config (pre-read from args so it
// can supply flag defaults) and returns a flag set that also accepts -config.
func newFlagSet(name string, args []string) (*flag.FlagSet, config, error) {
	cfg, err := loadConfig(flagValue(args, "config", defaultConfigFile))
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.String("config", defaultConfigFile, "YAML config file")
	return fs, cfg, err
}

// scenarioFlags defines the shared scenario selection flag; ids is its
// default (run.scenarios for the commands that play, nil for all). The
// scenarios themselves are embedded (package data).
func scenarioFlags(fs *flag.FlagSet, ids []string) loader {
	return loader{only: fs.String(constants.FlagOnly, strings.Join(ids, ","), "comma-separated scenario ids (default: run.scenarios in the config, empty for all)")}
}

func (l loader) load() ([]*scenario.Scenario, error) {
	scs, err := scenario.LoadAll(data.FS)
	if err != nil {
		return nil, err
	}
	if *l.only != "" {
		want := map[string]bool{}
		for _, id := range strings.Split(*l.only, ",") {
			want[strings.TrimSpace(id)] = true
		}
		var kept []*scenario.Scenario
		for _, s := range scs {
			if want[s.ID] {
				kept = append(kept, s)
			}
		}
		scs = kept
	}
	if len(scs) == 0 {
		return nil, fmt.Errorf("no scenarios selected")
	}
	return scs, nil
}

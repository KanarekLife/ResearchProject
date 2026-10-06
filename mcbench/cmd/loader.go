package main

import (
	"flag"
	"fmt"
	"strings"

	"mcbench/game/scenario"
)

// loader holds the scenario selection flags shared by most commands.
type loader struct{ dir, root, only *string }

// newFlagSet loads the config file named by -config (pre-read from args so it
// can supply flag defaults) and returns a flag set that also accepts -config.
func newFlagSet(name string, args []string) (*flag.FlagSet, config, error) {
	cfg, err := loadConfig(flagValue(args, "config", defaultConfigFile))
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.String("config", defaultConfigFile, "YAML config file")
	return fs, cfg, err
}

// scenarioFlags defines the shared scenario flags, with defaults from config.
func scenarioFlags(fs *flag.FlagSet, cfg config) loader {
	return loader{
		dir:  fs.String("scenarios", cfg.Data.Scenarios, "scenario directory"),
		root: fs.String("root", cfg.Data.Root, "directory holding decks/, villains/ and encounter-sets/"),
		only: fs.String("only", "", "comma-separated scenario ids (default: all)"),
	}
}

func (l loader) load() ([]*scenario.Scenario, error) {
	scs, err := scenario.LoadDir(*l.dir, *l.root)
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

package main

import (
	"flag"
	"fmt"
	"strings"

	"mcbench/game/scenario"
	"mcbench/game/session"
)

// loader holds the scenario selection flags shared by most commands.
type loader struct{ dir, root, only *string }

// scenarioFlags defines the shared scenario flags, with defaults from config.
// It also defines -config so the pre-read of the config file parses cleanly.
func scenarioFlags(fs *flag.FlagSet, cfg config) loader {
	fs.String("config", defaultConfigFile, "YAML config file")
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
		var kept []*scenario.Scenario
		for _, s := range scs {
			if strings.Contains(","+*l.only+",", ","+s.ID+",") {
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

// configFromArgs loads the config file named by -config (or config.yaml).
func configFromArgs(args []string) (config, error) {
	return loadConfig(flagValue(args, "config", defaultConfigFile))
}

func newSession(sc *scenario.Scenario, seed uint64) (*session.Session, error) {
	if seed == 0 {
		seed = sc.Seeds[0]
	}
	return session.New(sc, seed, session.Options{MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions})
}

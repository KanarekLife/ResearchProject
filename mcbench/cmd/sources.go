package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"mcbench/online"
)

func store(cfg config) online.Store {
	return online.Store{Repo: cfg.Online.Repo, Remote: cfg.Online.Remote, Branch: cfg.Online.Branch, Dir: cfg.Online.Dir}
}

func sourceFlag(fs *flag.FlagSet, cfg config) *string {
	return fs.String("source", cfg.Online.Source, "results to read: local (the results directory), online (the "+
		cfg.Online.Branch+" branch) or hybrid (both; a published local run is shown once)")
}

// resultDirs returns the directories a source reads, local first. Offline, the
// last synced checkout is read.
func resultDirs(source, local string, s online.Store) ([]string, error) {
	switch source {
	case sourceLocal:
		return []string{local}, nil
	case sourceOnline, sourceHybrid:
	default:
		return nil, fmt.Errorf("-source %q: want %s, %s or %s", source, sourceLocal, sourceOnline, sourceHybrid)
	}
	if err := s.Sync(); err != nil {
		if _, statErr := os.Stat(s.Dir); statErr != nil {
			return nil, err
		}
		slog.Warn("online results not updated; reading the last synced copy", "dir", s.Dir, "error", err)
	}
	if source == sourceOnline {
		return []string{s.Dir}, nil
	}
	return []string{local, s.Dir}, nil
}

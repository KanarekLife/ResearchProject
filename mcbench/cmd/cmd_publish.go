package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"mcbench/benchmark"
	"mcbench/constants"
)

const publishUsage = "usage: mcbench publish [-partial] RUN_DIR...\n"

// cmdPublish merges the games of runs of one player and publishes them as a
// single run. A game that ended in infra_error is dropped, or replaced by the
// same game in a later run, so a run and its retries publish together.
func cmdPublish(args []string) error {
	flags, cfg, err := newFlagSet(cmdNamePublish, args)
	if err != nil {
		return err
	}
	flags.Usage = func() {
		flags.Output().Write([]byte(publishUsage))
		flags.PrintDefaults()
	}
	partial := flags.Bool("partial", false, "publish even if fewer games finished than the first run planned")
	flags.Parse(args)
	dirs := flags.Args()
	if len(dirs) == 0 {
		flags.Usage()
		return fmt.Errorf("no run directory given")
	}
	tmp, err := os.MkdirTemp("", "mcbench-publish-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	merged, err := mergeRuns(dirs, tmp, *partial)
	if err != nil {
		return err
	}
	name, err := store(cfg).Publish(merged)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if err := os.WriteFile(filepath.Join(d, constants.PublishedFile), []byte(name+"\n"), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("published %d run(s) as %s on %s/%s\n", len(dirs), name, cfg.Online.Remote, cfg.Online.Branch)
	return nil
}

// mergeRuns writes the merged run into a directory under tmp, named like the
// first run, and returns it. Later runs win over earlier ones.
func mergeRuns(dirs []string, tmp string, partial bool) (string, error) {
	var sets [][]benchmark.Record
	for _, d := range dirs {
		recs, err := benchmark.ReadRecords(filepath.Join(d, gamesFile))
		if err != nil {
			return "", err
		}
		sets = append(sets, recs)
	}
	recs, from, err := mergeRecords(sets)
	if err != nil {
		return "", err
	}
	var info runInfo
	b, err := os.ReadFile(filepath.Join(dirs[0], constants.RunInfoFile))
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(b, &info); err != nil {
		return "", err
	}
	if len(recs) == 0 {
		return "", fmt.Errorf("no graded games left to publish: every game ended in infra_error")
	}
	if len(recs) < info.Planned && !partial {
		return "", fmt.Errorf("%d of %d planned games finished; pass -partial to publish them anyway", len(recs), info.Planned)
	}
	info.Statuses = map[string]int{}
	for _, r := range recs {
		info.Statuses[r.Status]++
	}
	out := filepath.Join(tmp, filepath.Base(dirs[0]))
	live := filepath.Join(out, liveSubdir)
	if err := os.MkdirAll(live, 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(filepath.Join(out, gamesFile))
	if err != nil {
		return "", err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for i, r := range recs {
		if err := enc.Encode(r); err != nil {
			return "", err
		}
		trace := r.GameName() + ".jsonl"
		if err := copyFile(filepath.Join(dirs[from[i]], liveSubdir, trace), filepath.Join(live, trace)); err != nil {
			return "", err
		}
	}
	b, err = json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", err
	}
	return out, os.WriteFile(filepath.Join(out, constants.RunInfoFile), append(b, '\n'), 0o644)
}

type gameKey struct {
	scenario string
	seed     uint64
	sample   int
}

// mergeRecords keeps one record per game: the latest that is not an
// infra_error. It returns them in game order with the index of the set each
// came from, and refuses sets of different players or instructions.
func mergeRecords(sets [][]benchmark.Record) ([]benchmark.Record, []int, error) {
	type pick struct {
		rec benchmark.Record
		set int
	}
	picked := map[gameKey]pick{}
	var first *benchmark.Record
	for i, set := range sets {
		for _, r := range set {
			if first == nil {
				first = &r
			}
			if r.Player != first.Player || r.Instructions.SHA256 != first.Instructions.SHA256 {
				return nil, nil, fmt.Errorf("runs differ: %s [%s] vs %s [%s]", first.Player, first.Instructions.SHA256, r.Player, r.Instructions.SHA256)
			}
			if r.Status != constants.InfraError {
				picked[gameKey{r.Scenario, r.Seed, r.Sample}] = pick{r, i}
			}
		}
	}
	keys := make([]gameKey, 0, len(picked))
	for k := range picked {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b gameKey) int {
		switch {
		case a.scenario != b.scenario:
			return strings.Compare(a.scenario, b.scenario)
		case a.seed != b.seed:
			return cmp.Compare(a.seed, b.seed)
		}
		return a.sample - b.sample
	})
	recs, from := make([]benchmark.Record, len(keys)), make([]int, len(keys))
	for i, k := range keys {
		recs[i], from[i] = picked[k].rec, picked[k].set
	}
	return recs, from, nil
}

func copyFile(from, to string) error {
	b, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, b, 0o644)
}

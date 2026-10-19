package main

import (
	"mcbench/benchmark"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()

	cfg, err := loadConfig(filepath.Join(dir, "missing.yaml"))
	if err != nil || cfg.Run.Seeds != defaultSeeds {
		t.Errorf("missing file: seeds=%d err=%v, want the defaults and no error", cfg.Run.Seeds, err)
	}

	path := filepath.Join(dir, "c.yaml")
	os.WriteFile(path, []byte("run:\n  seeds: 3\n"), 0o644)
	cfg, err = loadConfig(path)
	if err != nil || cfg.Run.Seeds != 3 || cfg.Run.Parallel != defaultParallel {
		t.Errorf("seeds=%d parallel=%d err=%v, want 3, the default parallel and no error", cfg.Run.Seeds, cfg.Run.Parallel, err)
	}

	os.WriteFile(path, []byte("run:\n  seedz: 3\n"), 0o644)
	if _, err = loadConfig(path); err == nil {
		t.Error("unknown config key accepted, want an error")
	}
}

func TestNewFlagSetReadsConfigBeforeFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(path, []byte("run:\n  seeds: 3\n"), 0o644)
	for _, args := range [][]string{{"-config", path}, {"--config=" + path}} {
		_, cfg, err := newFlagSet("x", args)
		if err != nil || cfg.Run.Seeds != 3 {
			t.Errorf("%v: seeds=%d err=%v, want 3", args, cfg.Run.Seeds, err)
		}
	}
}

func TestLoaderOnly(t *testing.T) {
	fs, _, _ := newFlagSet("x", nil)
	l := scenarioFlags(fs)
	fs.Parse([]string{"-only", "spider-man-vs-rhino"})
	if scs, err := l.load(); err != nil || len(scs) != 1 {
		t.Errorf("only existing id: %d scenarios, err=%v", len(scs), err)
	}
	fs.Parse([]string{"-only", "nope"})
	if _, err := l.load(); err == nil || !strings.Contains(err.Error(), "no scenarios") {
		t.Errorf("only unknown id: err=%v, want no scenarios selected", err)
	}
}

func TestResolveSelection(t *testing.T) {
	results := t.TempDir()
	run := filepath.Join(results, "20260101T000000Z-x")
	live := filepath.Join(run, "live")
	os.MkdirAll(live, 0o755)
	for _, n := range []string{"a_seed1_s0.jsonl", "a_seed2_s0.jsonl"} {
		os.WriteFile(filepath.Join(live, n), nil, 0o644)
	}
	for _, tc := range []struct {
		path    string
		results []string
		game    string
		want    string
	}{
		{filepath.Join(live, "a_seed1_s0.jsonl"), nil, "", "a_seed1_s0"},
		{"", []string{results}, "", ""}, // no game pinned: follow the run
		{run, nil, "seed1", "a_seed1_s0"},
	} {
		res, id, game, err := resolveSelection(tc.path, tc.results, "", tc.game)
		if err != nil || len(res) != 1 || res[0] != results || id != filepath.Base(run) || game != tc.want {
			t.Fatalf("resolveSelection(%q, %q, %q) = %q, %q, %q, %v; want game %q", tc.path, tc.results, tc.game, res, id, game, err, tc.want)
		}
	}
	if _, _, _, err := resolveSelection(filepath.Join(t.TempDir(), "g.jsonl"), nil, "", ""); err == nil {
		t.Fatal("a trace outside a run was accepted")
	}
}

// The label names only the reasoning setting each API is sent.
func TestSentReasoning(t *testing.T) {
	for _, tc := range []struct {
		api, effort string
		budget      int
		want        string
	}{
		{apiOpenAI, "medium", 2048, "m@effort-medium@budget-2048"},
		{apiGoogle, "medium", 2048, "m@effort-medium"},
		{apiGoogle, "", 2048, "m@budget-2048"},
		{apiAnthropic, "medium", 2048, "m@effort-medium"},
		{apiResponses, "", 2048, "m"},
	} {
		effort, budget := sentReasoning(tc.api, tc.effort, tc.budget)
		if got := modelLabel(runSettings{modelID: "m", effort: effort, budget: budget}); got != tc.want {
			t.Errorf("%s effort %q budget %d: label %q, want %q", tc.api, tc.effort, tc.budget, got, tc.want)
		}
	}
}

func TestResolveServer(t *testing.T) {
	t.Setenv(defaultServerEnv, "http://openai-gateway/v1")
	t.Setenv("MY_GATEWAY", "http://my-gateway/v1")
	cfg := func(server, env string) inferenceConfig { return inferenceConfig{Server: server, ServerEnv: env} }
	for _, tc := range []struct {
		api, flag string
		flagSet   bool
		cfg       inferenceConfig
		want      string
	}{
		{apiOpenAI, "", false, cfg(defaultServer, defaultServerEnv), "http://openai-gateway/v1"},
		{apiAnthropic, "", false, cfg(defaultServer, defaultServerEnv), "http://openai-gateway/v1"},
		{apiAnthropic, "", false, cfg(defaultServer, "MY_GATEWAY"), "http://my-gateway/v1"},
		{apiAnthropic, "", false, cfg(defaultServer, "UNSET_VAR"), ""},
		{apiGoogle, "", false, cfg("http://cfg/v1", "UNSET_VAR"), "http://cfg/v1"},
		{apiOpenAI, "", false, cfg(defaultServer, ""), defaultServer},
		{apiAnthropic, "http://flag/v1", true, cfg(defaultServer, "MY_GATEWAY"), "http://flag/v1"},
	} {
		if got := resolveServer(tc.api, tc.flag, tc.flagSet, tc.cfg); got != tc.want {
			t.Errorf("%s flag=%q cfg=%+v: got %q, want %q", tc.api, tc.flag, tc.cfg, got, tc.want)
		}
	}
}

func TestParseSeeds(t *testing.T) {
	if got, err := parseSeeds("1, 2,10"); err != nil || len(got) != 3 || got[2] != 10 {
		t.Errorf("got %v, %v", got, err)
	}
	if got, err := parseSeeds(""); err != nil || got != nil {
		t.Errorf("empty: got %v, %v", got, err)
	}
	if _, err := parseSeeds("1,x"); err == nil {
		t.Error("bad seed accepted")
	}
}

func TestMergeRecords(t *testing.T) {
	rec := func(seed uint64, status string, score float64) benchmark.Record {
		return benchmark.Record{Scenario: "s", Player: "p", Seed: seed, Status: status, Score: score}
	}
	recs, from, err := mergeRecords([][]benchmark.Record{
		{rec(1, "infra_error", 0), rec(2, "lost", 0.2), rec(3, "infra_error", 0)},
		{rec(1, "won", 0.9), rec(3, "infra_error", 0), rec(2, "lost", 0.3)},
	})
	if err != nil || len(recs) != 2 {
		t.Fatalf("got %d records, err=%v, want 2 (seed 3 stays failed and is dropped)", len(recs), err)
	}
	if recs[0].Seed != 1 || recs[0].Status != "won" || from[0] != 1 {
		t.Errorf("seed 1: %+v from set %d, want the retry's win", recs[0], from[0])
	}
	if recs[1].Seed != 2 || recs[1].Score != 0.3 || from[1] != 1 {
		t.Errorf("seed 2: %+v from set %d, want the later run's record", recs[1], from[1])
	}
	other := rec(1, "won", 1)
	other.Player = "q"
	if _, _, err := mergeRecords([][]benchmark.Record{{rec(1, "won", 1)}, {other}}); err == nil {
		t.Error("runs of different players merged, want an error")
	}
}

func TestMergeRunsSkipsEmpty(t *testing.T) {
	dir := t.TempDir()
	run := filepath.Join(dir, "r1")
	os.MkdirAll(run, 0o755)
	os.WriteFile(filepath.Join(run, gamesFile), []byte(`{"scenario":"s","player":"p","seed":1,"status":"infra_error"}`+"\n"), 0o644)
	os.WriteFile(filepath.Join(run, runInfoFile), []byte(`{"planned_games":1}`), 0o644)
	if _, err := mergeRuns([]string{run}, t.TempDir(), true); err == nil || !strings.Contains(err.Error(), "no graded games") {
		t.Errorf("err=%v, want a refusal to publish a run with no graded games", err)
	}
}

package main

import (
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
	fs, cfg, _ := newFlagSet("x", nil)
	cfg.Data.Scenarios, cfg.Data.Root = "../data/scenarios", "../data"
	l := scenarioFlags(fs, cfg)
	fs.Parse([]string{"-only", "spider-man-vs-rhino"})
	if scs, err := l.load(); err != nil || len(scs) != 1 {
		t.Errorf("only existing id: %d scenarios, err=%v", len(scs), err)
	}
	fs.Parse([]string{"-only", "nope"})
	if _, err := l.load(); err == nil || !strings.Contains(err.Error(), "no scenarios") {
		t.Errorf("only unknown id: err=%v, want no scenarios selected", err)
	}
}

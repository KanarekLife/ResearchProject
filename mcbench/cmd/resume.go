package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/player/instruction"
)

// resumeState is an earlier run that `run -resume` continues: its settings, and
// the games it finished, which are kept as they are. A game that ended with an
// infra_error is played again, from its trace when it has one.
type resumeState struct {
	dir      string
	info     runInfo
	finished []benchmark.Record
}

func loadResume(dir string) (*resumeState, error) {
	b, err := os.ReadFile(filepath.Join(dir, constants.RunInfoFile))
	if err != nil {
		return nil, fmt.Errorf("-resume: %w", err)
	}
	rs := &resumeState{dir: dir}
	if err := json.Unmarshal(b, &rs.info); err != nil {
		return nil, fmt.Errorf("-resume: %s: %w", constants.RunInfoFile, err)
	}
	recs, err := benchmark.ReadRecords(filepath.Join(dir, gamesFile))
	if err != nil {
		return nil, fmt.Errorf("-resume: %w", err)
	}
	for _, r := range recs {
		if r.Status != constants.InfraError {
			rs.finished = append(rs.finished, r)
		}
	}
	if rs.info.Settings.Player != playerModel {
		return nil, fmt.Errorf("-resume: only runs of -player %s can be continued", playerModel)
	}
	return rs, nil
}

// overlay makes the run's own settings the defaults, so that only what the
// command line names (a new -server, say) differs from the run being continued.
func (rs *resumeState) overlay(cfg config) config {
	s := rs.info.Settings
	cfg.Inference.API, cfg.Inference.Server, cfg.Inference.Model = s.API, s.Server, s.Model
	cfg.Inference.Temperature, cfg.Inference.MaxTokens, cfg.Inference.Retries = s.Temperature, s.MaxTokens, s.Retries
	cfg.Inference.ReasoningEffort, cfg.Inference.ReasoningBudget, cfg.Inference.Stream = s.ReasoningEffort, s.ReasoningBudget, s.Stream
	cfg.Player.Name, cfg.Player.MaxFailures, cfg.Player.Instructions = s.Player, s.MaxFailures, s.Instructions
	cfg.Run.Seeds, cfg.Run.Samples, cfg.Run.Parallel = s.Seeds, s.Samples, s.Parallel
	return cfg
}

// label is the name the run's games carry, kept even when the model is reached
// under another id.
func (rs *resumeState) label() string {
	s := rs.info.Settings
	return modelLabel(runSettings{modelID: s.Model, effort: s.ReasoningEffort, budget: s.ReasoningBudget})
}

// check refuses a continuation that would no longer be the same benchmark run.
func (rs *resumeState) check(instr instruction.Instructions) error {
	for _, r := range rs.finished {
		if r.Instructions.SHA256 != instr.SHA256 {
			return fmt.Errorf("-resume: the instruction documents changed since this run (hash %s, was %s)", instr.SHA256, r.Instructions.SHA256)
		}
	}
	if commit := gitLine("rev-parse", "HEAD"); commit != rs.info.Commit {
		slog.Warn("resuming on another commit than the run started on", "run", rs.info.Commit, "now", commit)
	}
	return nil
}

// reopen rewrites games.jsonl with the finished games and opens it to append to.
func (rs *resumeState) reopen() (*os.File, benchmark.Config, error) {
	f, err := os.Create(filepath.Join(rs.dir, gamesFile))
	if err != nil {
		return nil, benchmark.Config{}, err
	}
	enc := json.NewEncoder(f)
	done := map[string]bool{}
	for _, r := range rs.finished {
		if err := enc.Encode(r); err != nil {
			f.Close()
			return nil, benchmark.Config{}, err
		}
		done[r.GameName()] = true
	}
	return f, benchmark.Config{TraceDir: filepath.Join(rs.dir, liveSubdir), RunName: filepath.Base(rs.dir), Resume: true, Done: done}, nil
}

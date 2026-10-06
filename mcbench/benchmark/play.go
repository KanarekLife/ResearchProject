package benchmark

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Play plays one game and scores it. With a traceDir it also writes a JSON
// trace to <dir>/<scenario>_seed<N>_s<K>.jsonl as the game happens.
func Play(ctx context.Context, sc *scenario.Scenario, p player.Player, instr instruction.Instructions, seed uint64, sample int, traceDir string, weights Weights) Record {
	start := time.Now()
	rec := Record{Scenario: sc.ID, Player: p.Name(), Instructions: instr, Seed: seed, Sample: sample}
	log := slog.Default().With("scenario", sc.ID, "seed", seed, "sample", sample)

	opts := session.Options{Logger: log, MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions, ShuffleOptions: true, Sample: uint64(sample)}
	traceFile, err := attachTrace(&opts, traceDir, sc.ID, seed, sample)
	if err != nil {
		rec.Status, rec.Error = constants.AgentError, err.Error()
		return rec
	}
	if traceFile != nil {
		defer traceFile.Close()
	}

	s, err := session.New(sc, seed, opts)
	if err != nil {
		rec.Status, rec.Error = constants.EngineError, err.Error()
		return rec
	}
	traceStart(s, &rec)
	log.Info("game started", "player", p.Name())

	usage, err := p.Play(ctx, s)
	rec.Usage = usage
	rec.Status, rec.Error = outcome(s, err)

	finalize(&rec, s, sc.MaxRounds, weights)
	rec.Millis = time.Since(start).Milliseconds()
	traceEnd(s, &rec)
	if rec.Error != "" {
		log.Warn("game ended abnormally", "status", rec.Status, "error", rec.Error)
	}
	return rec
}

type startTrace struct {
	session.TraceHeader
	Scenario     string                   `json:"scenario"`
	Seed         uint64                   `json:"seed"`
	Sample       int                      `json:"sample"`
	Player       string                   `json:"player"`
	Instructions instruction.Instructions `json:"instructions"`
}

type endTrace struct {
	session.TraceHeader
	Status   string             `json:"status"`
	Error    string             `json:"error"`
	Rounds   int                `json:"rounds"`
	Criteria map[string]float64 `json:"criteria"`
	Score    float64            `json:"score"`
	Usage    player.Usage       `json:"usage"`
}

// outcome is the record status and error of a finished game: the session's
// status, unless the player itself failed.
func outcome(s *session.Session, playErr error) (status, errMsg string) {
	status = s.Status()
	switch {
	case status == constants.EngineError:
		return status, s.Error()
	case playErr != nil:
		return constants.AgentError, playErr.Error()
	}
	return status, ""
}

func traceStart(s *session.Session, rec *Record) {
	s.Trace(constants.TraceStart, &startTrace{
		Scenario: rec.Scenario, Seed: rec.Seed, Sample: rec.Sample, Player: rec.Player, Instructions: rec.Instructions,
	})
}

func traceEnd(s *session.Session, rec *Record) {
	s.Trace(constants.TraceEnd, &endTrace{
		Status: rec.Status, Error: rec.Error, Rounds: rec.Rounds, Criteria: rec.Criteria, Score: rec.Score, Usage: rec.Usage,
	})
}

func finalize(rec *Record, s *session.Session, maxRounds int, weights Weights) {
	rec.Stats = s.Stats()
	rec.Rounds = int(rec.Stats[constants.MetricRound])
	rec.Decisions = len(s.Choices)
	rec.Invalid = s.Invalid
	rec.Choices = s.Choices
	rec.Log = s.Log(0)
	rec.Criteria = Criteria(rec.Stats, maxRounds)
	rec.Score = Score(rec.Criteria, weights)
	if rec.Status == constants.AgentError || rec.Status == constants.EngineError {
		rec.Score = 0 // a broken game never earns credit
	}
}

// attachTrace opens the per-game trace file and points the session at it.
func attachTrace(opts *session.Options, dir, scenarioName string, seed uint64, sample int) (*os.File, error) {
	if dir == "" {
		return nil, nil
	}
	f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s_seed%d_s%d.jsonl", scenarioName, seed, sample)))
	if err != nil {
		return nil, err
	}
	opts.Trace = f
	return f, nil
}

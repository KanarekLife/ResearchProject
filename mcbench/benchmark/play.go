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

// Play plays one game and scores it. With a traceDir it also writes a
// readable board to <dir>/<scenario>_seed<N>_s<K>.txt and a JSON trace to
// <...>.jsonl, both as the game happens.
func Play(ctx context.Context, sc *scenario.Scenario, p player.Player, instr instruction.Instructions, seed uint64, sample int, traceDir string, weights Weights) Record {
	start := time.Now()
	rec := Record{Scenario: sc.ID, Player: p.Name(), Instructions: instr, Seed: seed, Sample: sample}
	log := slog.Default().With("scenario", sc.ID, "seed", seed, "sample", sample)

	opts := session.Options{Logger: log, MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions, ShuffleOptions: true, Sample: uint64(sample)}
	files, err := attachTrace(&opts, traceDir, sc.ID, p.Name(), seed, sample)
	if err != nil {
		rec.Status, rec.Error = constants.AgentError, err.Error()
		return rec
	}
	defer closeFiles(files)

	s, err := session.New(sc, seed, opts)
	if err != nil {
		rec.Status, rec.Error = session.EngineError, err.Error()
		return rec
	}
	traceStart(s, &rec)
	log.Info("game started", "player", p.Name())

	usage, err := p.Play(ctx, s)
	rec.Status, rec.Usage = s.Status(), usage
	switch {
	case rec.Status == session.EngineError:
		rec.Error = s.Error()
	case err != nil:
		rec.Status, rec.Error = constants.AgentError, err.Error()
	}

	finalize(&rec, s, sc.MaxRounds, weights)
	rec.Millis = time.Since(start).Milliseconds()
	traceEnd(s, &rec)
	if rec.Error != "" {
		log.Warn("game ended abnormally", "status", rec.Status, "error", rec.Error)
	}
	return rec
}

func traceStart(s *session.Session, rec *Record) {
	s.Trace(map[string]any{
		constants.TraceKeyType:     constants.TraceStart,
		constants.TraceKeyScenario: rec.Scenario,
		constants.TraceKeySeed:     rec.Seed,
		constants.TraceKeySample:   rec.Sample,
		constants.TraceKeyPlayer:   rec.Player,
		constants.TraceKeyInstr:    rec.Instructions,
	})
}

func traceEnd(s *session.Session, rec *Record) {
	s.Trace(map[string]any{
		constants.TraceKeyType:     constants.TraceEnd,
		constants.TraceKeyStatus:   rec.Status,
		constants.TraceKeyError:    rec.Error,
		constants.TraceKeyRounds:   rec.Rounds,
		constants.TraceKeyCriteria: rec.Criteria,
		constants.TraceKeyScore:    rec.Score,
		constants.TraceKeyUsage:    rec.Usage,
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
	if rec.Status == constants.AgentError || rec.Status == session.EngineError {
		rec.Score = 0 // a broken game never earns credit
	}
}

// attachTrace opens the per-game trace files and points the session at them.
func attachTrace(opts *session.Options, dir, scenarioName, playerName string, seed uint64, sample int) ([]*os.File, error) {
	if dir == "" {
		return nil, nil
	}
	base := fmt.Sprintf("%s_seed%d_s%d", scenarioName, seed, sample)
	trace, err := os.Create(filepath.Join(dir, base+".jsonl"))
	if err != nil {
		return nil, err
	}
	live, err := os.Create(filepath.Join(dir, base+".txt"))
	if err != nil {
		trace.Close()
		return nil, err
	}
	fmt.Fprintf(live, "scenario %s · seed %d · sample %d · player %s\n", scenarioName, seed, sample, playerName)
	opts.Trace, opts.Live = trace, live
	return []*os.File{trace, live}, nil
}

func closeFiles(files []*os.File) {
	for _, f := range files {
		f.Close()
	}
}

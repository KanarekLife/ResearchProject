// Package bench is the benchmark harness: it plays scenario games with an
// agent, scores them and writes records.
package bench

import (
	"context"
	"sync"
	"time"

	"mcbench/agent"
	"mcbench/game"
	"mcbench/scenario"
)

// Record is one game.
type Record struct {
	RunID        string             `json:"run_id"`
	Scenario     string             `json:"scenario"`
	Agent        string             `json:"agent"`
	Instructions agent.Instructions `json:"instructions"`
	Seed         uint64             `json:"seed"`
	Sample       int                `json:"sample"`
	Status       string             `json:"status"` // a game status, or "agent_error"
	Error        string             `json:"error,omitempty"`
	Rounds       int                `json:"rounds"`
	Decisions    int                `json:"decisions"`
	Invalid      int                `json:"invalid_choices,omitempty"`
	Stats        map[string]float64 `json:"stats"`
	Criteria     map[string]float64 `json:"criteria"`
	Score        float64            `json:"score"`
	Usage        agent.Usage        `json:"usage,omitzero"`
	Millis       int64              `json:"ms"`
	Choices      []game.Choice      `json:"choices"`
	Log          []string           `json:"log"`
}

// PlayGame plays one game of a scenario on a seed and scores it.
func PlayGame(ctx context.Context, sc *scenario.Scenario, ag agent.Agent, instr agent.Instructions, seed uint64, sample int) Record {
	start := time.Now()
	rec := Record{Scenario: sc.ID, Agent: ag.Name(), Instructions: instr, Seed: seed, Sample: sample}
	s, err := game.New(sc, seed, game.Options{
		MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions, ShuffleOptions: true, Sample: uint64(sample),
	})
	if err != nil {
		rec.Status, rec.Error = game.EngineError, err.Error()
		return rec
	}
	usage, err := ag.Play(ctx, s)
	rec.Status, rec.Usage = s.Status(), usage
	if rec.Status == game.EngineError {
		rec.Error = s.Error()
	} else if err != nil {
		rec.Status, rec.Error = "agent_error", err.Error()
	}
	rec.Stats = s.Stats()
	rec.Rounds = int(rec.Stats["round"])
	rec.Decisions = len(s.Choices)
	rec.Invalid = s.Invalid
	rec.Choices = s.Choices
	rec.Log = s.Log(0)
	rec.Criteria = Criteria(rec.Stats, sc.MaxRounds)
	rec.Score = Score(rec.Criteria, sc.Scoring)
	if rec.Status == "agent_error" || rec.Status == game.EngineError {
		rec.Score = 0 // a broken game never earns credit
	}
	rec.Millis = time.Since(start).Milliseconds()
	return rec
}

// Config of a benchmark run.
type Config struct {
	Seeds    int // first N seeds of each scenario (0 = all)
	Samples  int // games per seed
	Parallel int
}

type job struct {
	sc     *scenario.Scenario
	seed   uint64
	sample int
}

func jobs(scs []*scenario.Scenario, cfg Config) []job {
	var out []job
	for _, sc := range scs {
		seeds := sc.Seeds
		if cfg.Seeds > 0 && cfg.Seeds < len(seeds) {
			seeds = seeds[:cfg.Seeds]
		}
		for _, seed := range seeds {
			for i := 0; i < max(1, cfg.Samples); i++ {
				out = append(out, job{sc, seed, i})
			}
		}
	}
	return out
}

// Jobs counts the games a Run would play.
func Jobs(scs []*scenario.Scenario, cfg Config) int { return len(jobs(scs, cfg)) }

// Run plays every scenario on its seeds and emits each finished game.
func Run(ctx context.Context, scs []*scenario.Scenario, ag agent.Agent, instr agent.Instructions, cfg Config, emit func(Record)) {
	js := jobs(scs, cfg)
	runID := time.Now().UTC().Format("20060102T150405Z")
	next := make(chan job)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < max(1, cfg.Parallel); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range next {
				rec := PlayGame(ctx, j.sc, ag, instr, j.seed, j.sample)
				rec.RunID = runID
				mu.Lock()
				emit(rec)
				mu.Unlock()
			}
		}()
	}
	for _, j := range js {
		if ctx.Err() != nil {
			break
		}
		next <- j
	}
	close(next)
	wg.Wait()
}

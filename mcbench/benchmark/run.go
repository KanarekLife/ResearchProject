package benchmark

import (
	"context"
	"sync"
	"time"

	"mcbench/game/scenario"
	"mcbench/player"
	"mcbench/player/instruction"
)

// Config of a run.
type Config struct {
	Seeds    int // first N seeds of each scenario (0 = all)
	Samples  int // games per seed
	Parallel int // concurrent games
	TraceDir string
}

type job struct {
	sc     *scenario.Scenario
	seed   uint64
	sample int
}

func Count(scenarios []*scenario.Scenario, cfg Config) int { return len(plan(scenarios, cfg)) }

func plan(scenarios []*scenario.Scenario, cfg Config) []job {
	var out []job
	for _, sc := range scenarios {
		seeds := sc.Seeds
		if cfg.Seeds > 0 && cfg.Seeds < len(seeds) {
			seeds = seeds[:cfg.Seeds]
		}
		for _, seed := range seeds {
			for sample := 0; sample < max(1, cfg.Samples); sample++ {
				out = append(out, job{sc, seed, sample})
			}
		}
	}
	return out
}

// Run plays every game and calls emit once per finished game, in order. emit
// runs on the caller's goroutine, so it need not be concurrency-safe.
func Run(ctx context.Context, scenarios []*scenario.Scenario, p player.Player, instr instruction.Instructions, cfg Config, weights Weights, emit func(Record)) {
	jobs := plan(scenarios, cfg)
	runID := time.Now().UTC().Format("20060102T150405Z")

	work := make(chan job)
	results := make(chan Record)
	var wg sync.WaitGroup
	for w := 0; w < max(1, cfg.Parallel); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range work {
				rec := Play(ctx, j.sc, p, instr, j.seed, j.sample, cfg.TraceDir, weights)
				rec.RunID = runID
				results <- rec
			}
		}()
	}
	go func() {
		for _, j := range jobs {
			if ctx.Err() != nil {
				break
			}
			work <- j
		}
		close(work)
		wg.Wait()
		close(results)
	}()

	for rec := range results {
		emit(rec)
	}
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"mcbench/benchmark"
	"mcbench/game/scenario"
	"mcbench/player"
	"mcbench/player/instruction"
)

// runSettings are the parsed flags of `mcbench run`.
type runSettings struct {
	player      string
	modelID     string
	api         string
	server      string
	temp        float64
	maxTokens   int
	retries     int
	effort      string
	budget      int
	maxFailures int
	rag         bool
	vision      bool
	instrDir    string
	instr       []string
	seeds       int
	samples     int
	parallel    int
	logLevel    string
	out         string
}

func cmdRun(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameRun, args)
	if err != nil {
		return err
	}
	l, o := parseRunFlags(fs, args, cfg)
	if err := setupLogging(o.logLevel); err != nil {
		return err
	}
	scs, err := l.load()
	if err != nil {
		return err
	}
	instr, err := instruction.LoadInstructions(o.instrDir, o.instr)
	if err != nil {
		return err
	}
	if o.player == playerModel && o.modelID == "" {
		return fmt.Errorf("-model is required for -player model")
	}
	client, err := newClient(o, os.Getenv(cfg.Inference.APIKeyEnv))
	if err != nil {
		return err
	}
	p, err := buildPlayer(o, client, instr)
	if err != nil {
		return err
	}
	return executeRun(scs, p, instr, o, cfg.Scoring)
}

func parseRunFlags(fs *flag.FlagSet, args []string, cfg config) (loader, runSettings) {
	l := scenarioFlags(fs)
	playerName := fs.String("player", cfg.Player.Name, "player: model, heuristic, random, first")
	modelID := fs.String("model", cfg.Inference.Model, "model id (required for -player model)")
	api := fs.String("api", cfg.Inference.API, "API protocol: openai (chat completions), responses (OpenAI Responses), google (Gemini generateContent), anthropic (Messages)")
	server := fs.String("server", envOr("OPENAI_BASE_URL", cfg.Inference.Server), "API base URL (OpenAI-compatible: local LM Studio by default; google and anthropic: the provider's own URL unless set here)")
	temp := fs.Float64("temperature", cfg.Inference.Temperature, "sampling temperature; negative (default -1): not sent, the provider default applies")
	maxTokens := fs.Int("max-tokens", cfg.Inference.MaxTokens, "max completion tokens per request (reasoning models need room)")
	retries := fs.Int("retries", cfg.Inference.Retries, "retry transient API failures")
	effort := fs.String("reasoning-effort", cfg.Inference.ReasoningEffort, "reasoning level for hosted models, e.g. low, medium, high (empty: provider default)")
	budget := fs.Int("reasoning-budget", cfg.Inference.ReasoningBudget, "max thinking tokens for llama.cpp-style servers, e.g. 4096 (0: server default)")
	maxFailures := fs.Int("max-failures", cfg.Player.MaxFailures, "failed model replies (no tool call or a tool error) allowed per decision")
	rag := fs.Bool("rag", false, "POC: BM25 passages from the rules reference PDF per decision, not the instruction documents in the system prompt")
	vision := fs.Bool("vision", false, "POC: a picture of the board with each decision (openai API only)")
	instrDir := fs.String("instructions-dir", cfg.Player.InstructionsDir, "instruction documents directory")
	instrNames := fs.String("instructions", strings.Join(cfg.Player.Instructions, ","), "comma-separated instruction documents (empty for none)")
	seeds := fs.Int("seeds", cfg.Run.Seeds, "play only the first N seeds of each scenario (0 = all)")
	samples := fs.Int("samples", cfg.Run.Samples, "games per seed (more than 1 whenever sampling is not deterministic)")
	parallel := fs.Int("parallel", cfg.Run.Parallel, "concurrent games")
	out := fs.String("out", cfg.Run.Out, "results directory")
	logLevel := fs.String("log-level", "info", logLevelUsage)
	fs.Parse(args)
	if *api != apiOpenAI {
		// Only an explicit -server or a configured server applies; otherwise the
		// provider's own URL does, never the OpenAI-compatible default.
		serverSet := false
		fs.Visit(func(f *flag.Flag) { serverSet = serverSet || f.Name == "server" })
		if !serverSet && cfg.Inference.Server == defaultServer {
			*server = ""
		} else if !serverSet {
			*server = cfg.Inference.Server
		}
	}

	*effort, *budget = sentReasoning(*api, *effort, *budget)

	var instr []string
	if *instrNames != "" {
		instr = strings.Split(*instrNames, ",")
	}
	return l, runSettings{
		player: *playerName, modelID: *modelID, api: *api, server: *server,
		temp: *temp, maxTokens: *maxTokens, retries: *retries, effort: *effort, budget: *budget, maxFailures: *maxFailures,
		rag: *rag, vision: *vision,
		instrDir: *instrDir, instr: instr,
		seeds: *seeds, samples: *samples, parallel: *parallel, out: *out, logLevel: *logLevel,
	}
}

func executeRun(scs []*scenario.Scenario, p player.Player, instr instruction.Instructions, o runSettings, scoring map[string]float64) error {
	f, runCfg, err := prepareRunDir(o, p)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := slog.Default()
	enc := json.NewEncoder(f)
	total, done := benchmark.Count(scs, runCfg), 0
	var recs []benchmark.Record
	log.Info("run started", "player", p.Name(), "games", total, "parallel", o.parallel, "live_dir", runCfg.TraceDir)
	benchmark.Run(ctx, scs, p, instr, runCfg, benchmark.Weights(scoring), func(r benchmark.Record) {
		done++
		enc.Encode(r)
		recs = append(recs, r)
		log.Info("game finished",
			"done", done, "total", total, "scenario", r.Scenario, "seed", r.Seed,
			"status", r.Status, "rounds", r.Rounds, "decisions", r.Decisions,
			"score", r.Score, "ms", r.Millis, "error", firstLines(r.Error, 1))
	})
	log.Info("run finished", "games_file", f.Name())
	benchmark.Report(os.Stdout, recs, terminalTheme(os.Stdout))
	return nil
}

// prepareRunDir creates results/<run>-<player>[__<instructions>]/, the live/
// trace directory and the games.jsonl file. The instruction documents are in
// the name so runs of one model that differ only in them stay apart.
func prepareRunDir(o runSettings, p player.Player) (*os.File, benchmark.Config, error) {
	name := time.Now().UTC().Format(runIDLayout) + "-" + sanitize(p.Name())
	if len(o.instr) > 0 {
		name += runNameSep + sanitize(strings.Join(o.instr, "+"))
	}
	dir := filepath.Join(o.out, name)
	runCfg := benchmark.Config{Seeds: o.seeds, Samples: o.samples, Parallel: o.parallel, TraceDir: filepath.Join(dir, liveSubdir)}
	if err := os.MkdirAll(runCfg.TraceDir, 0o755); err != nil {
		return nil, runCfg, err
	}
	f, err := os.Create(filepath.Join(dir, gamesFile))
	if err != nil {
		return nil, runCfg, err
	}
	return f, runCfg, nil
}

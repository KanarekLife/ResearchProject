package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/online"
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
	headers     map[string]string
	effort      string
	budget      int
	maxFailures int
	instrDir    string
	instr       []string
	seeds       int
	samples     int
	parallel    int
	logLevel    string
	out         string
	publish     bool
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
	return executeRun(scs, p, instr, o, cfg.Scoring, store(cfg))
}

func parseRunFlags(fs *flag.FlagSet, args []string, cfg config) (loader, runSettings) {
	l := scenarioFlags(fs)
	playerName := fs.String("player", cfg.Player.Name, "player: model, heuristic, random, first")
	modelID := fs.String("model", cfg.Inference.Model, "model id (required for -player model)")
	api := fs.String("api", cfg.Inference.API, "API protocol: openai (chat completions), responses (OpenAI Responses), google (Gemini generateContent), anthropic (Messages)")
	server := fs.String("server", cfg.Inference.Server, "API base URL (default: the protocol's environment variable, then the config; openai is local LM Studio by default, the other protocols use their provider's own URL)")
	temp := fs.Float64("temperature", cfg.Inference.Temperature, "sampling temperature; negative (default -1): not sent, the provider default applies")
	maxTokens := fs.Int("max-tokens", cfg.Inference.MaxTokens, "max completion tokens per request (reasoning models need room)")
	retries := fs.Int("retries", cfg.Inference.Retries, "retry transient API failures")
	headers := headerFlag{}
	fs.Var(headers, "header", "extra request header \"Name: value\"; repeatable; {run} and {game} in the value are filled per game")
	effort := fs.String("reasoning-effort", cfg.Inference.ReasoningEffort, "reasoning level for hosted models, e.g. low, medium, high (empty: provider default)")
	budget := fs.Int("reasoning-budget", cfg.Inference.ReasoningBudget, "max thinking tokens for llama.cpp-style servers, e.g. 4096 (0: server default)")
	maxFailures := fs.Int("max-failures", cfg.Player.MaxFailures, "failed model replies (no tool call or a tool error) allowed per decision")
	instrDir := fs.String("instructions-dir", cfg.Player.InstructionsDir, "instruction documents directory")
	instrNames := fs.String("instructions", strings.Join(cfg.Player.Instructions, ","), "comma-separated instruction documents (empty for none)")
	seeds := fs.Int("seeds", cfg.Run.Seeds, "play only the first N seeds of each scenario (0 = all)")
	samples := fs.Int("samples", cfg.Run.Samples, "games per seed (more than 1 whenever sampling is not deterministic)")
	parallel := fs.Int("parallel", cfg.Run.Parallel, "concurrent games")
	out := fs.String("out", cfg.Run.Out, "results directory")
	publish := fs.Bool("publish", cfg.Online.Publish, "publish the run to the "+cfg.Online.Branch+" branch once it finishes with no infra_error game")
	logLevel := fs.String("log-level", "info", logLevelUsage)
	fs.Parse(args)
	serverSet := false
	fs.Visit(func(f *flag.Flag) { serverSet = serverSet || f.Name == "server" })
	*server = resolveServer(*api, *server, serverSet, cfg.Inference.Server)

	*effort, *budget = sentReasoning(*api, *effort, *budget)

	var instr []string
	if *instrNames != "" {
		instr = strings.Split(*instrNames, ",")
	}
	return l, runSettings{
		player: *playerName, modelID: *modelID, api: *api, server: *server,
		temp: *temp, maxTokens: *maxTokens, retries: *retries, headers: headers, effort: *effort, budget: *budget, maxFailures: *maxFailures,
		instrDir: *instrDir, instr: instr,
		seeds: *seeds, samples: *samples, parallel: *parallel, out: *out, logLevel: *logLevel, publish: *publish,
	}
}

func executeRun(scs []*scenario.Scenario, p player.Player, instr instruction.Instructions, o runSettings, scoring map[string]float64, st online.Store) error {
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
	info := startInfo(scs, o, scoring, total)
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
	dir := filepath.Dir(f.Name())
	if err := info.finish(dir, recs); err != nil {
		return err
	}
	if o.publish {
		publishRun(dir, info, st)
	}
	return nil
}

// publishRun skips a run cut short or with an infra_error game: it is incomplete.
func publishRun(dir string, info runInfo, st online.Store) {
	log := slog.Default().With("run", filepath.Base(dir))
	finished := 0
	for _, n := range info.Statuses {
		finished += n
	}
	switch {
	case finished < info.Planned:
		log.Warn("run not published: cut short", "finished", finished, "planned", info.Planned)
		return
	case info.Statuses[constants.InfraError] > 0:
		log.Warn("run not published: flaky", "infra_errors", info.Statuses[constants.InfraError])
		return
	}
	name, err := st.Publish(dir)
	if err != nil {
		log.Error("run not published", "error", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, constants.PublishedFile), []byte(name+"\n"), 0o644); err != nil {
		log.Error("run published, but recording that failed", "error", err)
	}
	log.Info("run published", "remote", st.Remote, "branch", st.Branch, "dir", name)
}

// prepareRunDir creates results/<run>-<player>[__<instructions>]/, the live/
// trace directory and the games.jsonl file. The instruction documents are in
// the name so runs of one model that differ only in them stay apart; a run
// started in the same second as one of the same name gets a -2, -3, ... suffix.
func prepareRunDir(o runSettings, p player.Player) (*os.File, benchmark.Config, error) {
	name := time.Now().UTC().Format(runIDLayout) + "-" + sanitize(p.Name())
	if len(o.instr) > 0 {
		name += runNameSep + sanitize(strings.Join(o.instr, "+"))
	}
	dir, err := newRunDir(o.out, name)
	runCfg := benchmark.Config{Seeds: o.seeds, Samples: o.samples, Parallel: o.parallel, TraceDir: filepath.Join(dir, liveSubdir), RunName: filepath.Base(dir)}
	if err != nil {
		return nil, runCfg, err
	}
	if err := os.Mkdir(runCfg.TraceDir, 0o755); err != nil {
		return nil, runCfg, err
	}
	f, err := os.Create(filepath.Join(dir, gamesFile))
	if err != nil {
		return nil, runCfg, err
	}
	return f, runCfg, nil
}

func newRunDir(out, name string) (string, error) {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return "", err
	}
	dir := filepath.Join(out, name)
	for n := 2; ; n++ {
		err := os.Mkdir(dir, 0o755)
		if !errors.Is(err, os.ErrExist) {
			return dir, err
		}
		dir = filepath.Join(out, name+"-"+strconv.Itoa(n))
	}
}

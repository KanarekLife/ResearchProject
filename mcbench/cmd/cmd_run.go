package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/integrations/inference"
	"mcbench/online"
	"mcbench/player"
	"mcbench/player/instruction"
)

// runSettings are the parsed flags of `mcbench run`.
type runSettings struct {
	player        string
	modelID       string
	api           string
	server        string
	temp          float64
	maxTokens     int
	retries       int
	rateLimitWait time.Duration
	headers       map[string]string
	effort        string
	budget        int
	maxFailures   int
	instrDir      string
	instr         []string
	seeds         int
	pick          string
	samples       int
	parallel      int
	logLevel      string
	out           string
	publish       bool
	recentEvents  int
	stream        bool
	resume        string // run directory to continue
	label         string // the model label of a resumed run
	job           string // the job file this run plays, if any
	provider      string // the config provider the endpoint came from, if any
}

func cmdRun(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return playRun(ctx, args)
}

// playRun is `run` until ctx ends, which stops the games being played.
func playRun(ctx context.Context, args []string) error {
	fs, cfg, err := newFlagSet(cmdNameRun, args)
	if err != nil {
		return err
	}
	var rs *resumeState
	if dir := flagValue(args, constants.FlagResume, ""); dir != "" {
		if rs, err = loadResume(dir); err != nil {
			return err
		}
		cfg = rs.overlay(cfg)
	}
	provider := flagValue(args, constants.FlagProvider, "")
	if provider == "" && rs != nil {
		provider = rs.info.Settings.Provider
	}
	if provider != "" {
		if cfg, err = cfg.withProvider(provider); err != nil {
			return err
		}
	}
	l, o := parseRunFlags(fs, args, cfg)
	o.provider = provider
	if rs != nil {
		o.resume, o.label = rs.dir, rs.label()
	}
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
	if (o.player == playerModel || o.player == playerSystemOne) && o.modelID == "" {
		return fmt.Errorf("-model is required for -player %s", o.player)
	}
	if rs != nil {
		if err := rs.check(instr); err != nil {
			return err
		}
	}
	apiKey := os.Getenv(cfg.Inference.APIKeyEnv)
	var client inference.Client
	if o.player != playerSystemOne {
		if client, err = newClient(o, apiKey); err != nil {
			return err
		}
	}
	p, err := buildPlayer(o, client, systemOneClient(o, apiKey), instr)
	if err != nil {
		return err
	}
	return executeRun(ctx, scs, p, instr, o, cfg, rs)
}

func parseRunFlags(fs *flag.FlagSet, args []string, cfg config) (loader, runSettings) {
	l := scenarioFlags(fs, cfg.Run.Scenarios)
	playerName := fs.String(constants.FlagPlayer, cfg.Player.Name, "player: model, systemone (a System One classifier, see -model and -server), heuristic, random, first")
	modelID := fs.String(constants.FlagModel, cfg.Inference.Model, "model id (required for -player model)")
	api := fs.String("api", cfg.Inference.API, "API protocol: openai (chat completions), responses (OpenAI Responses), google (Gemini generateContent), anthropic (Messages)")
	server := fs.String("server", cfg.Inference.Server, "API base URL (default: the server_env variable, then the config; openai is local LM Studio by default, the other protocols use their provider's own URL)")
	temp := fs.Float64(constants.FlagTemperature, cfg.Inference.Temperature, "sampling temperature; negative (default -1): not sent, the provider default applies")
	maxTokens := fs.Int(constants.FlagMaxTokens, cfg.Inference.MaxTokens, "max completion tokens per request (reasoning models need room)")
	retries := fs.Int("retries", cfg.Inference.Retries, "retry transient API failures")
	rateLimitWait := fs.Duration("rate-limit-wait", cfg.Inference.RateLimitWait, "how long to wait out rate limits (HTTP 429) per request, e.g. 6h; they do not spend -retries (0: retry them like other failures)")
	fs.String(constants.FlagProvider, "", "a provider of the config: its api, server, key, headers, stream, retries and rate-limit wait become the defaults")
	headers := headerFlag{}
	maps.Copy(headers, cfg.Inference.Headers)
	fs.Var(headers, "header", "extra request header \"Name: value\"; repeatable; {run} and {game} in the value are filled per game")
	effort := fs.String(constants.FlagReasoningEffort, cfg.Inference.ReasoningEffort, "reasoning level for hosted models, e.g. low, medium, high (empty: provider default)")
	budget := fs.Int(constants.FlagReasoningBudget, cfg.Inference.ReasoningBudget, "max thinking tokens for llama.cpp-style servers, e.g. 4096 (0: server default)")
	stream := fs.Bool("stream", cfg.Inference.Stream, "stream the replies (every -api), for gateways that cut a slow request after a couple of minutes")
	maxFailures := fs.Int(constants.FlagMaxFailures, cfg.Player.MaxFailures, "failed model replies (no tool call or a tool error) allowed per decision")
	instrDir := fs.String("instructions-dir", cfg.Player.InstructionsDir, "instruction documents directory")
	instrNames := fs.String(constants.FlagInstructions, strings.Join(cfg.Player.Instructions, ","), "comma-separated instruction documents (empty for none; none by default for -player systemone, which takes rules_short)")
	seeds := fs.Int(constants.FlagSeeds, cfg.Run.Seeds, "play only the first N seeds of each scenario (0 = all)")
	pick := fs.String(constants.FlagPick, "", "comma-separated seeds to play, e.g. to repeat games that hit an infra error (default: all)")
	samples := fs.Int(constants.FlagSamples, cfg.Run.Samples, "games per seed (more than 1 whenever sampling is not deterministic)")
	parallel := fs.Int(constants.FlagParallel, cfg.Run.Parallel, "concurrent games")
	out := fs.String("out", cfg.Run.Out, "results directory")
	publish := fs.Bool("publish", cfg.Online.Publish, "publish the run to the "+cfg.Online.Branch+" branch once it finishes with no infra_error game")
	recentEvents := fs.Int("recent-events", defaultRecentEvents, "systemone player: log lines sent with each decision (0: the whole log)")
	job := fs.String(constants.FlagJob, "", "the job this run belongs to, recorded in run.json (set by mcbench jobs)")
	fs.String(constants.FlagResume, "", "continue the run in this results directory: its finished games are kept, the others continue from their traces (the run's own settings apply unless a flag changes them, e.g. -server)")
	logLevel := fs.String("log-level", "info", logLevelUsage)
	fs.Parse(args)
	serverSet, instrSet := false, false
	fs.Visit(func(f *flag.Flag) {
		serverSet = serverSet || f.Name == "server"
		instrSet = instrSet || f.Name == constants.FlagInstructions
	})
	*server = resolveServer(*api, *server, serverSet, cfg.Inference)

	*effort, *budget = sentReasoning(*api, *effort, *budget)

	var instr []string
	// The System One classifier cannot take the long documents, so it gets
	// none unless asked.
	if *playerName == playerSystemOne && !instrSet {
		*instrNames = ""
	}
	if *instrNames != "" {
		instr = strings.Split(*instrNames, ",")
	}
	return l, runSettings{
		player: *playerName, modelID: *modelID, api: *api, server: *server,
		temp: *temp, maxTokens: *maxTokens, retries: *retries, rateLimitWait: *rateLimitWait, headers: headers, effort: *effort, budget: *budget, stream: *stream, maxFailures: *maxFailures,
		instrDir: *instrDir, instr: instr,
		seeds: *seeds, pick: *pick, samples: *samples, parallel: *parallel, out: *out, logLevel: *logLevel, publish: *publish, recentEvents: *recentEvents, job: *job,
	}
}

func executeRun(ctx context.Context, scs []*scenario.Scenario, p player.Player, instr instruction.Instructions, o runSettings, cfg config, rs *resumeState) error {
	var (
		f      *os.File
		runCfg benchmark.Config
		err    error
	)
	if rs != nil {
		if f, runCfg, err = rs.reopen(); err == nil {
			runCfg.Seeds, runCfg.Samples, runCfg.Parallel = o.seeds, o.samples, o.parallel
			runCfg.Pick, err = parseSeeds(o.pick)
		}
	} else {
		f, runCfg, err = prepareRunDir(o, p)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	log := slog.Default()
	enc := json.NewEncoder(f)
	total, done := benchmark.Count(scs, runCfg), 0
	info := startInfo(scs, o, cfg.Scoring, total)
	var recs []benchmark.Record
	if rs != nil {
		info, recs = rs.info, slices.Clone(rs.finished)
		info.Statuses = map[string]int{}
		info.Resumed = append(info.Resumed, resumeInfo{Time: time.Now().UTC(), Server: o.server, Commit: headCommit()})
	}
	if err := info.write(filepath.Dir(f.Name())); err != nil {
		return err
	}
	log.Info("run started", "player", p.Name(), "games", total, "parallel", o.parallel, "live_dir", runCfg.TraceDir)
	benchmark.Run(ctx, scs, p, instr, runCfg, benchmark.Weights(cfg.Scoring), func(r benchmark.Record) {
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
	// A job the worker gave up on is finished once its resumed run is published.
	if o.publish && publishRun(dir, info, store(cfg)) && rs != nil && info.Job != "" {
		if err := jobsBranch(cfg).ArchiveFailed(info.Job); err != nil {
			log.Error("job not archived", "job", info.Job, "error", err)
		}
	}
	return nil
}

// publishRun skips a run cut short or with an infra_error game: it is incomplete.
func publishRun(dir string, info runInfo, st online.Store) bool {
	log := slog.Default().With("run", filepath.Base(dir))
	finished := 0
	for _, n := range info.Statuses {
		finished += n
	}
	switch {
	case finished < info.Planned:
		log.Warn("run not published: cut short", "finished", finished, "planned", info.Planned)
		return false
	case info.Statuses[constants.InfraError] > 0:
		log.Warn("run not published: flaky", "infra_errors", info.Statuses[constants.InfraError])
		return false
	}
	name, err := st.Publish(dir)
	if err != nil {
		log.Error("run not published", "error", err)
		return false
	}
	if err := os.WriteFile(filepath.Join(dir, constants.PublishedFile), []byte(name+"\n"), 0o644); err != nil {
		log.Error("run published, but recording that failed", "error", err)
	}
	log.Info("run published", "remote", st.Remote, "branch", st.Branch, "dir", name)
	return true
}

// prepareRunDir creates results/<run>-<player>[__<instructions>]/, the live/
// trace directory and the games.jsonl file. The instruction documents are in
// the name so runs of one model that differ only in them stay apart; a run
// started in the same second as one of the same name gets a -2, -3, ... suffix.
func prepareRunDir(o runSettings, p player.Player) (*os.File, benchmark.Config, error) {
	pick, err := parseSeeds(o.pick)
	if err != nil {
		return nil, benchmark.Config{}, err
	}
	name := time.Now().UTC().Format(runIDLayout) + "-" + sanitize(p.Name())
	if len(o.instr) > 0 {
		name += runNameSep + sanitize(strings.Join(o.instr, "+"))
	}
	dir, err := newRunDir(o.out, name)
	runCfg := benchmark.Config{Seeds: o.seeds, Pick: pick, Samples: o.samples, Parallel: o.parallel, TraceDir: filepath.Join(dir, liveSubdir), RunName: filepath.Base(dir)}
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

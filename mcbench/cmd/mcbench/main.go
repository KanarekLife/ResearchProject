// Command mcbench runs the Marvel Champions full-game benchmark.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"mcbench/agent"
	"mcbench/bench"
	"mcbench/game"
	"mcbench/scenario"
)

const usage = `mcbench - Marvel Champions full-game benchmark

Commands:
  list      list scenarios
  tools     print the agent <-> game tool contract (JSON schemas)
  show      print the first get_state and get_decision of a scenario game
  validate  play every scenario seed with the baselines (engine smoke test)
  play      play a scenario yourself in the terminal
  run       play scenarios with an agent and write game records
  report    summarize games.jsonl files

Run "mcbench <command> -h" for flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmds := map[string]func([]string) error{
		"list": cmdList, "tools": cmdTools, "show": cmdShow, "validate": cmdValidate,
		"play": cmdPlay, "run": cmdRun, "report": cmdReport,
	}
	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err := cmd(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// loader holds the scenario selection flags shared by most commands.
type loader struct{ dir, root, only *string }

func scenarioFlags(fs *flag.FlagSet) loader {
	return loader{
		dir:  fs.String("scenarios", "scenarios", "scenario directory"),
		root: fs.String("root", ".", "directory holding decks/, villains/ and encounter-sets/"),
		only: fs.String("only", "", "comma-separated scenario ids (default: all)"),
	}
}

func (l loader) load() ([]*scenario.Scenario, error) {
	scs, err := scenario.LoadDir(*l.dir, *l.root)
	if err != nil {
		return nil, err
	}
	if *l.only != "" {
		var kept []*scenario.Scenario
		for _, s := range scs {
			if strings.Contains(","+*l.only+",", ","+s.ID+",") {
				kept = append(kept, s)
			}
		}
		scs = kept
	}
	if len(scs) == 0 {
		return nil, fmt.Errorf("no scenarios selected")
	}
	return scs, nil
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	l := scenarioFlags(fs)
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	for _, s := range scs {
		fmt.Printf("%s: %s (%d seeds, max %d rounds)\n", s.ID, s.Title, len(s.Seeds), s.MaxRounds)
	}
	return nil
}

func cmdTools([]string) error {
	out, err := json.MarshalIndent(game.Tools, "", "  ")
	fmt.Println(string(out))
	return err
}

func cmdShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	l := scenarioFlags(fs)
	seed := fs.Uint64("seed", 0, "seed (default: the scenario's first)")
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	s, err := newSession(scs[0], *seed)
	if err != nil {
		return err
	}
	fmt.Printf("get_state:\n%s\n\nget_decision:\n%s\n", s.Call("get_state", nil), s.Call("get_decision", nil))
	return nil
}

func newSession(sc *scenario.Scenario, seed uint64) (*game.Session, error) {
	if seed == 0 {
		seed = sc.Seeds[0]
	}
	return game.New(sc, seed, game.Options{MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions})
}

// cmdValidate plays every seed with the baselines. Careless play reaches
// many engine paths, so any engine error here is a bug.
func cmdValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	l := scenarioFlags(fs)
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	failed := false
	for _, ag := range []agent.Agent{agent.Random{Seed: 1}, agent.First{}, agent.Heuristic{}} {
		var recs []bench.Record
		bench.Run(context.Background(), scs, ag, agent.Instructions{}, bench.Config{Parallel: 4}, func(r bench.Record) {
			recs = append(recs, r)
			if r.Status == game.EngineError || r.Status == "agent_error" || r.Status == game.DecisionLimit {
				failed = true
				fmt.Printf("%s seed %d (%s): %s\n  %s\n", r.Scenario, r.Seed, r.Agent, r.Status, firstLines(r.Error, 8))
			}
		})
		bench.Report(os.Stdout, recs)
	}
	if failed {
		return fmt.Errorf("engine errors found")
	}
	return nil
}

func cmdPlay(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	l := scenarioFlags(fs)
	seed := fs.Uint64("seed", 0, "seed (default: the scenario's first)")
	fs.Parse(args)
	scs, err := l.load()
	if err != nil {
		return err
	}
	if *seed == 0 {
		*seed = scs[0].Seeds[0]
	}
	rec := bench.PlayGame(context.Background(), scs[0], &agent.Human{In: bufio.NewReader(os.Stdin), Out: os.Stdout}, agent.Instructions{}, *seed, 0, "")
	out, _ := json.MarshalIndent(map[string]any{"status": rec.Status, "rounds": rec.Rounds, "criteria": rec.Criteria, "score": rec.Score}, "", "  ")
	fmt.Println(string(out))
	return nil
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	l := scenarioFlags(fs)
	agentName := fs.String("agent", "llm", "agent: llm, random, first, heuristic")
	baseURL := fs.String("base-url", envOr("OPENAI_BASE_URL", "http://localhost:1234/v1"), "OpenAI-compatible API base URL (local LM Studio by default)")
	model := fs.String("model", "", "model id (required for -agent llm)")
	protocol := fs.String("protocol", agent.ProtocolJSON, "json (harness fetches state; any model) or tools (model calls the game tools)")
	maxSteps := fs.Int("max-steps", 5, "model requests allowed per decision")
	temp := fs.Float64("temperature", 0, "sampling temperature")
	maxTokens := fs.Int("max-tokens", 32768, "max completion tokens per request (reasoning models need room)")
	instrDir := fs.String("instructions-dir", "instructions", "instruction documents directory")
	instrNames := fs.String("instructions", "rules", "comma-separated instruction documents (empty for none)")
	seeds := fs.Int("seeds", 0, "play only the first N seeds of each scenario (0 = all)")
	samples := fs.Int("samples", 1, "games per seed (use with temperature > 0)")
	parallel := fs.Int("parallel", 1, "concurrent games")
	out := fs.String("out", "results", "results directory")
	fs.Parse(args)

	scs, err := l.load()
	if err != nil {
		return err
	}
	var names []string
	if *instrNames != "" {
		names = strings.Split(*instrNames, ",")
	}
	instr, err := agent.LoadInstructions(*instrDir, names)
	if err != nil {
		return err
	}
	var ag agent.Agent
	switch *agentName {
	case "llm":
		if *model == "" {
			return fmt.Errorf("-model is required for -agent llm")
		}
		if *protocol != agent.ProtocolJSON && *protocol != agent.ProtocolTools {
			return fmt.Errorf("-protocol must be json or tools")
		}
		ag = &agent.LLM{
			Chat:     agent.Chat{BaseURL: *baseURL, APIKey: os.Getenv("OPENAI_API_KEY"), Model: *model, Temperature: *temp, MaxTokens: *maxTokens},
			Protocol: *protocol, MaxSteps: *maxSteps, Instructions: instr,
		}
	case "random":
		ag = agent.Random{Seed: 1}
	case "first":
		ag = agent.First{}
	case "heuristic":
		ag = agent.Heuristic{}
	default:
		return fmt.Errorf("unknown agent %q", *agentName)
	}

	dir := filepath.Join(*out, time.Now().UTC().Format("20060102T150405Z")+"-"+strings.NewReplacer("/", "_", ":", "_").Replace(ag.Name()))
	cfg := bench.Config{Seeds: *seeds, Samples: *samples, Parallel: *parallel, TraceDir: filepath.Join(dir, "live")}
	if err := os.MkdirAll(cfg.TraceDir, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "live traces: %s\n", cfg.TraceDir)
	f, err := os.Create(filepath.Join(dir, "games.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	total, done := bench.Jobs(scs, cfg), 0
	var recs []bench.Record
	bench.Run(ctx, scs, ag, instr, cfg, func(r bench.Record) {
		done++
		enc.Encode(r)
		recs = append(recs, r)
		fmt.Fprintf(os.Stderr, "[%d/%d] %s seed %d: %s after %d rounds (%d decisions, %.0fs), score %.3f %s\n",
			done, total, r.Scenario, r.Seed, r.Status, r.Rounds, r.Decisions, float64(r.Millis)/1000, r.Score, firstLines(r.Error, 1))
	})
	fmt.Fprintf(os.Stderr, "\ngames: %s\n", f.Name())
	bench.Report(os.Stdout, recs)
	return nil
}

func cmdReport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcbench report GAMES.jsonl...")
	}
	var all []bench.Record
	for _, p := range args {
		recs, err := bench.ReadRecords(p)
		if err != nil {
			return err
		}
		all = append(all, recs...)
	}
	bench.Report(os.Stdout, all)
	return nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	return strings.Join(lines[:min(n, len(lines))], "\n  ")
}

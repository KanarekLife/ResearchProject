package main

import (
	"cmp"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mcbench/constants"
	planusage "mcbench/integrations/usage"
	"mcbench/jobs"
	"mcbench/online"
	"mcbench/trace"
)

func jobsBranch(cfg config) jobs.Branch {
	return jobs.Branch{Store: online.Store{Repo: cfg.Online.Repo, Remote: cfg.Online.Remote, Branch: cfg.Jobs.Branch, Dir: cfg.Jobs.Dir}}
}

func cmdJobs(args []string) error {
	fs, cfg, err := newFlagSet(cmdNameJobs, args)
	if err != nil {
		return err
	}
	watch := fs.Duration("watch", 0, "poll the "+cfg.Jobs.Branch+" branch at this interval (0: play the pending jobs once and exit)")
	fs.Parse(args)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	host, _ := os.Hostname()
	w := jobs.Worker{
		Branch:    jobsBranch(cfg),
		Host:      host,
		Args:      []string{"-config", flagValue(args, "config", defaultConfigFile), "-publish=true"},
		Play:      cmdRun,
		Ready:     jobReady(ctx, cfg),
		LatestRun: func(job string, since time.Time) *jobs.Run { return latestJobRun(cfg.Run.Out, job, since) },
	}
	w.Watch(ctx, *watch)
	return nil
}

// jobReady takes a job only while every usage window of its provider's plan
// is below jobs.max_usage. A plan that cannot be read does not hold a job
// back: the clients still wait out a rate limit.
func jobReady(ctx context.Context, cfg config) func(name string, data []byte) bool {
	return func(name string, data []byte) bool {
		j, err := jobs.Parse(data)
		if err != nil {
			return true // fails when played, with its reason in errors/
		}
		p := cfg.Providers[j.Model.Provider]
		if p.Usage == "" {
			return true
		}
		keyEnv := cmp.Or(p.APIKeyEnv, cfg.Inference.APIKeyEnv)
		windows, err := planusage.Check(ctx, p.Usage, os.Getenv(keyEnv))
		if err != nil {
			slog.Warn("provider usage not read, taking the job", "job", name, "provider", j.Model.Provider, "error", err)
			return true
		}
		for _, w := range windows {
			if w.Percent >= cfg.Jobs.MaxUsage {
				slog.Info("job waits for its provider's usage window", "job", name, "provider", j.Model.Provider,
					"window", w.Name, "percent", w.Percent, "max", cfg.Jobs.MaxUsage, "resets", w.ResetsAt)
				return false
			}
		}
		return true
	}
}

// latestJobRun is the newest run of a job in out started at since or later.
func latestJobRun(out, job string, since time.Time) *jobs.Run {
	entries, _ := os.ReadDir(out)
	var latest *jobs.Run
	for _, e := range entries {
		dir := filepath.Join(out, e.Name())
		b, err := os.ReadFile(filepath.Join(dir, constants.RunInfoFile))
		if err != nil {
			continue
		}
		var info runInfo
		if json.Unmarshal(b, &info) == nil && info.Job == job && !info.Started.Before(since) {
			latest = &jobs.Run{Dir: dir, Statuses: info.Statuses, Planned: info.Planned, Published: trace.Published(dir) != ""}
		}
	}
	return latest
}

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mcbench/constants"
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
		LatestRun: func(job string, since time.Time) *jobs.Run { return latestJobRun(cfg.Run.Out, job, since) },
	}
	w.Watch(ctx, *watch)
	return nil
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

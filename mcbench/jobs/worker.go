package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"mcbench/constants"
)

// Watch settles the jobs this worker left in running/, then plays the pending
// jobs every interval until ctx ends; with no interval, it plays them once.
func (w Worker) Watch(ctx context.Context, interval time.Duration) {
	log := slog.With("worker", w.Host, "branch", w.Branch.Store.Branch)
	log.Info("worker started", "interval", interval)
	w.recover()
	for {
		played := 0
		for ctx.Err() == nil && w.next() {
			played++
		}
		if ctx.Err() != nil {
			log.Info("worker stopped", "reason", context.Cause(ctx))
			return
		}
		log.Info("no job pending", "played", played)
		if interval == 0 {
			return
		}
		select {
		case <-ctx.Done():
		case <-time.After(interval):
		}
	}
}

// next claims the first pending job and plays it; false when none was claimed.
func (w Worker) next() bool {
	c, err := w.Branch.Claim(w.Host)
	if err != nil {
		slog.Error("jobs branch not read", "error", err)
		return false
	}
	if c == nil {
		return false
	}
	slog.Info("job claimed", "job", c.Name)
	w.play(c.Name, c.Data)
	return true
}

// play runs a claimed job as `run` would, always publishing, and archives it
// once its run is published.
func (w Worker) play(name string, data []byte) {
	log := slog.With("job", name)
	args, err := w.runArgs(name, data)
	if err != nil {
		w.fail(failure{job: name, reason: fmt.Sprintf("job file: %v", err)})
		return
	}
	log.Info("job started", "args", strings.Join(args, " "))
	start := time.Now().UTC()
	err = w.Play(args)
	run := w.LatestRun(name, start)
	log.Info("job run ended", "elapsed", time.Since(start).Round(time.Second), "error", err, "run", runDir(run))
	f := failure{job: name, run: run, args: args}
	switch {
	case err != nil:
		f.reason = err.Error()
	case run == nil:
		f.reason = "the run wrote no run directory"
	case !run.Published:
		f.reason = "run not published: cut short, an infra_error game or a failed push (see the worker log)"
	default:
		w.archive(name)
		return
	}
	w.fail(f)
}

// recover settles the jobs this worker left in running/ when it stopped: its
// results volume holds their runs. A published run's job is archived, any
// other is moved to errors/.
func (w Worker) recover() {
	jobs, err := w.Branch.Running()
	if err != nil {
		slog.Error("jobs branch not read", "error", err)
		return
	}
	for name, data := range jobs {
		run := w.LatestRun(name, time.Time{})
		if run == nil {
			slog.Info("running job left alone: no local run", "job", name)
			continue // another worker's, or stopped before its run started
		}
		slog.Info("recovering job", "job", name, "run", run.Dir, "published", run.Published)
		if run.Published {
			w.archive(name)
			continue
		}
		args, _ := w.runArgs(name, data)
		w.fail(failure{job: name, reason: "the worker stopped while the job ran", run: run, args: args})
	}
}

// runArgs are the `run` flags that play a job.
func (w Worker) runArgs(name string, data []byte) ([]string, error) {
	j, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if j.Name != "" {
		slog.Info("job read", "job", name, "name", j.Name)
	}
	args := append(append([]string{}, w.Args...), j.Args()...)
	return append(args, "-"+constants.FlagJob+"="+name), nil
}

func (w Worker) archive(name string) {
	if err := w.Branch.Archive(name); err != nil {
		slog.Error("job not archived", "job", name, "error", err)
		return
	}
	slog.Info("job finished", "job", name)
}

func (w Worker) fail(f failure) {
	f.worker, f.time = w.Host, time.Now().UTC()
	slog.Error("job failed", "job", f.job, "error", f.reason, "run", runDir(f.run))
	if err := w.Branch.Fail(f.job, f.String()); err != nil {
		slog.Error("job failure not recorded", "job", f.job, "error", err)
		return
	}
	slog.Info("job moved to errors", "job", f.job)
}

func (f failure) String() string {
	s := fmt.Sprintf("job: %s\nworker: %s\ntime: %s\nerror: %s\n", f.job, f.worker, f.time.Format(time.RFC3339), f.reason)
	if f.run != nil {
		s += fmt.Sprintf("run: %s\nstatuses: %v of %d planned games\nresume: mcbench run %s -%s %s\n",
			f.run.Dir, f.run.Statuses, f.run.Planned, strings.Join(f.args, " "), constants.FlagResume, f.run.Dir)
	}
	return s
}

func runDir(r *Run) string {
	if r == nil {
		return ""
	}
	return r.Dir
}

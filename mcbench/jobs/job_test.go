package jobs

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"mcbench/online"
)

// A job sets only the flags of the fields it names, so the rest keeps the
// worker's default, and it cannot set the endpoint, which is the provider's.
func TestJobSetsOnlyItsFields(t *testing.T) {
	j, err := Parse([]byte("model:\n  provider: proxy\n  id: m\ninstructions: []\nseeds: 2\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-provider=proxy", "-model=m", "-instructions=", "-seeds=2"}
	if got := j.Args(); !slices.Equal(got, want) {
		t.Errorf("args %q, want %q", got, want)
	}
	if _, err := Parse([]byte("model:\n  server: http://x\n")); err == nil {
		t.Error("a job set the server, which belongs to the worker's provider")
	}
}

// Once a worker claims a job, another finds nothing pending; .yml and .yaml
// files are both jobs.
func TestJobClaimedOnce(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	gitT(t, "", "init", "--quiet", "--bare", remote)
	var branches []Branch
	for _, name := range []string{"a", "b"} {
		repo := filepath.Join(root, name)
		gitT(t, "", "clone", "--quiet", remote, repo)
		gitT(t, repo, "config", "user.name", "test")
		gitT(t, repo, "config", "user.email", "test@example.com")
		branches = append(branches, Branch{Store: online.Store{Repo: repo, Remote: "origin", Branch: "jobs", Dir: filepath.Join(root, name+"-jobs")}})
	}
	err := branches[0].Store.Commit(func(dir string) (string, error) {
		os.MkdirAll(filepath.Join(dir, pendingDir), 0o755)
		os.WriteFile(filepath.Join(dir, pendingDir, "j.yml"), []byte("seeds: 1\n"), 0o644)
		return "add jobs", os.WriteFile(filepath.Join(dir, pendingDir, "k.yaml"), []byte("seeds: 1\n"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"j", "k", ""} {
		c, err := branches[i%2].Claim("test")
		got := ""
		if c != nil {
			got = c.Name
		}
		if err != nil || got != want {
			t.Fatalf("claim %d got %q (%v), want %q", i, got, err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(branches[0].Store.Dir, runningDir, "j.yml")); err != nil {
		t.Errorf("claimed job not in %s/: %v", runningDir, err)
	}
}

// With two slots, two pending jobs play at the same time, each publishes its
// run through the one results checkout, and both are archived.
func TestWorkerPlaysJobsConcurrently(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	gitT(t, "", "init", "--quiet", "--bare", remote)
	repo := filepath.Join(root, "repo")
	gitT(t, "", "clone", "--quiet", remote, repo)
	gitT(t, repo, "config", "user.name", "test")
	gitT(t, repo, "config", "user.email", "test@example.com")
	b := Branch{Store: online.Store{Repo: repo, Remote: "origin", Branch: "jobs", Dir: filepath.Join(root, "jobs")}}
	results := online.Store{Repo: repo, Remote: "origin", Branch: "results", Dir: filepath.Join(root, "results")}
	err := b.Store.Commit(func(dir string) (string, error) {
		os.MkdirAll(filepath.Join(dir, pendingDir), 0o755)
		os.WriteFile(filepath.Join(dir, pendingDir, "j.yaml"), []byte("seeds: 1\n"), 0o644)
		return "add jobs", os.WriteFile(filepath.Join(dir, pendingDir, "k.yaml"), []byte("seeds: 1\n"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	published := map[string]bool{}
	var started sync.WaitGroup
	started.Add(2)
	w := Worker{Branch: b, Host: "test", Concurrent: 2,
		// Each job waits for the other to start, so they only finish if they
		// play at the same time.
		Play: func(args []string) error {
			job := strings.TrimPrefix(args[len(args)-1], "-job=")
			started.Done()
			if !waitTimeout(&started, 10*time.Second) {
				t.Errorf("job %s played alone", job)
			}
			run := filepath.Join(root, "local", job)
			os.MkdirAll(run, 0o755)
			if err := os.WriteFile(filepath.Join(run, "games.jsonl"), []byte("{}\n"), 0o644); err != nil {
				return err
			}
			if _, err := results.Publish(run); err != nil {
				return err
			}
			mu.Lock()
			published[job] = true
			mu.Unlock()
			return nil
		},
		LatestRun: func(job string, _ time.Time) *Run {
			mu.Lock()
			defer mu.Unlock()
			return &Run{Dir: job, Published: published[job]}
		},
	}
	w.Watch(context.Background(), 0)

	if err := results.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := b.Store.Sync(); err != nil {
		t.Fatal(err)
	}
	for _, job := range []string{"j", "k"} {
		if _, err := os.Stat(filepath.Join(results.Dir, job, "games.jsonl")); err != nil {
			t.Errorf("run of %s not published: %v", job, err)
		}
		if _, err := os.Stat(filepath.Join(b.Store.Dir, archiveDir, job+".yaml")); err != nil {
			t.Errorf("job %s not archived: %v", job, err)
		}
	}
}

// waitTimeout waits for wg, false when it took longer than d.
func waitTimeout(wg *sync.WaitGroup, d time.Duration) bool {
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		return true
	case <-time.After(d):
		return false
	}
}

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

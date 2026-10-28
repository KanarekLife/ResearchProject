package jobs

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

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
	branches := pendingJobs(t, 2)
	for i, want := range []string{"j", "k", ""} {
		c, err := branches[i%2].Claim("test", nil)
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

// A job its provider has no room for stays pending while a later one is taken.
func TestClaimSkipsJobsNotReady(t *testing.T) {
	b := pendingJobs(t, 1)[0]
	notJ := func(name string, _ []byte) bool { return name != "j" }
	for _, step := range []struct {
		ready func(string, []byte) bool
		want  string
	}{{notJ, "k"}, {notJ, ""}, {nil, "j"}} {
		c, err := b.Claim("test", step.ready)
		got := ""
		if c != nil {
			got = c.Name
		}
		if err != nil || got != step.want {
			t.Fatalf("claimed %q (%v), want %q", got, err, step.want)
		}
	}
}

// pendingJobs gives n workers' checkouts of one jobs branch with the pending
// jobs j (.yml) and k (.yaml).
func pendingJobs(t *testing.T, n int) []Branch {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	gitT(t, "", "init", "--quiet", "--bare", remote)
	var branches []Branch
	for i := range n {
		repo := filepath.Join(root, string(rune('a'+i)))
		gitT(t, "", "clone", "--quiet", remote, repo)
		gitT(t, repo, "config", "user.name", "test")
		gitT(t, repo, "config", "user.email", "test@example.com")
		branches = append(branches, Branch{Store: online.Store{Repo: repo, Remote: "origin", Branch: "jobs", Dir: repo + "-jobs"}})
	}
	err := branches[0].Store.Commit(func(dir string) (string, error) {
		os.MkdirAll(filepath.Join(dir, pendingDir), 0o755)
		os.WriteFile(filepath.Join(dir, pendingDir, "j.yml"), []byte("seeds: 1\n"), 0o644)
		return "add jobs", os.WriteFile(filepath.Join(dir, pendingDir, "k.yaml"), []byte("seeds: 1\n"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return branches
}

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

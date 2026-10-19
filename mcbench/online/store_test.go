package online

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPublishKeepsCollidingRuns(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	mustGit(t, "", "init", "--quiet", "--bare", remote)
	var stores []Store
	for _, name := range []string{"a", "b"} {
		repo := filepath.Join(root, name)
		mustGit(t, "", "clone", "--quiet", remote, repo)
		mustGit(t, repo, "config", "user.name", "test")
		mustGit(t, repo, "config", "user.email", "test@example.com")
		stores = append(stores, Store{Repo: repo, Remote: "origin", Branch: "results", Dir: filepath.Join(root, name+"-results")})
	}

	run := filepath.Join(root, "local", "20260101T000000Z-model_x")
	os.MkdirAll(filepath.Join(run, "live"), 0o755)
	os.WriteFile(filepath.Join(run, "games.jsonl"), []byte("{}\n"), 0o644)
	os.WriteFile(filepath.Join(run, "live", "g.jsonl"), []byte("{}\n"), 0o644)

	var got []string
	for _, s := range stores {
		name, err := s.Publish(run)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, name)
	}
	if got[0] != "20260101T000000Z-model_x" || got[1] != "20260101T000000Z-model_x-2" {
		t.Fatalf("published as %v, want the run's name, then it with -2", got)
	}
	if err := stores[0].Sync(); err != nil {
		t.Fatal(err)
	}
	for _, name := range got {
		if _, err := os.Stat(filepath.Join(stores[0].Dir, name, "live", "g.jsonl")); err != nil {
			t.Errorf("%s missing from the synced branch: %v", name, err)
		}
	}
}

func TestSyncRefusesABranchCheckout(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	mustGit(t, "", "init", "--quiet", "--bare", remote)
	repo := filepath.Join(root, "repo")
	mustGit(t, "", "clone", "--quiet", remote, repo)
	if err := (Store{Repo: repo, Remote: "origin", Branch: "results", Dir: repo}).Sync(); err == nil {
		t.Fatal("Sync accepted the repository's own working tree")
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := git(dir, args...); err != nil {
		t.Fatal(err)
	}
}

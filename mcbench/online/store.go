// Package online commits to git branches kept apart from the code, such as the
// published runs and the job files, through a local checkout of each.
package online

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Store is a results branch on a remote and its local checkout.
type Store struct {
	Repo   string // empty: the current directory's repository
	Remote string
	Branch string
	Dir    string
}

const publishAttempts = 5

// missingRefExit is what `git ls-remote --exit-code` returns for a missing ref.
const missingRefExit = 2

// Sync brings the checkout to the remote branch, creating both on first use.
func (s Store) Sync() error {
	exists, err := s.remoteHasBranch()
	if err != nil {
		return err
	}
	if exists {
		if _, err := git(s.Repo, "fetch", "--quiet", s.Remote, "+refs/heads/"+s.Branch+":"+s.trackingRef()); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.Dir); errors.Is(err, os.ErrNotExist) {
		start := s.trackingRef()
		if !exists {
			if start, err = s.emptyRoot(); err != nil {
				return err
			}
		}
		_, err := git(s.Repo, "worktree", "add", "--quiet", "--detach", s.Dir, start)
		return err
	}
	if err := s.checkWorktree(); err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if _, err := git(s.Dir, "reset", "--quiet", "--hard", s.trackingRef()); err != nil {
		return err
	}
	_, err = git(s.Dir, "clean", "--quiet", "-fd")
	return err
}

// Publish pushes runDir to the branch as a new top-level directory and returns
// its name there, suffixed -2, -3, ... when taken.
func (s Store) Publish(runDir string) (string, error) {
	var name string
	err := s.Commit(func(dir string) (string, error) {
		name = freeName(dir, filepath.Base(runDir))
		return "chore(results): add " + name, os.CopyFS(filepath.Join(dir, name), os.DirFS(runDir))
	})
	return name, err
}

// Commit runs change on a freshly synced checkout and pushes what it did as one
// commit with the message it returns; an empty message commits nothing. A lost
// push race is retried on the new branch tip, an error from change is not.
func (s Store) Commit(change func(dir string) (msg string, err error)) error {
	var err error
	for range publishAttempts {
		if err = s.Sync(); err != nil {
			continue
		}
		msg, changeErr := change(s.Dir)
		if changeErr != nil || msg == "" {
			return changeErr
		}
		if err = s.push(msg); err == nil {
			return nil
		}
	}
	return err
}

func (s Store) push(msg string) error {
	steps := [][]string{
		{"add", "--all"},
		{"commit", "--quiet", "-m", msg},
		{"push", "--quiet", s.Remote, "HEAD:refs/heads/" + s.Branch},
	}
	for _, args := range steps {
		if _, err := git(s.Dir, args...); err != nil {
			return err
		}
	}
	return nil
}

func freeName(dir, base string) string {
	name := base
	for n := 2; ; n++ {
		if _, err := os.Stat(filepath.Join(dir, name)); errors.Is(err, os.ErrNotExist) {
			return name
		}
		name = base + "-" + strconv.Itoa(n)
	}
}

func (s Store) trackingRef() string { return "refs/remotes/" + s.Remote + "/" + s.Branch }

func (s Store) remoteHasBranch() (bool, error) {
	_, err := git(s.Repo, "ls-remote", "--quiet", "--exit-code", "--heads", s.Remote, s.Branch)
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == missingRefExit {
		return false, nil
	}
	return err == nil, err
}

// checkWorktree keeps Sync's hard reset off any tree on a branch, such as the
// repository's own: a results checkout is always detached.
func (s Store) checkWorktree() error {
	top, err := git(s.Dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(s.Dir)
	if err != nil {
		return err
	}
	same, _ := filepath.EvalSymlinks(abs)
	if top != same {
		return fmt.Errorf("%s is not the root of a results checkout (its root is %s)", s.Dir, top)
	}
	if _, err := git(s.Dir, "symbolic-ref", "--quiet", "HEAD"); err == nil {
		return fmt.Errorf("%s is on a branch; a results checkout is always detached", s.Dir)
	}
	return nil
}

func (s Store) emptyRoot() (string, error) {
	tree, err := git(s.Repo, "mktree")
	if err != nil {
		return "", err
	}
	return git(s.Repo, "commit-tree", tree, "-m", "chore("+s.Branch+"): start the "+s.Branch+" branch")
}

// git returns the trimmed output. Its stdin is empty, which mktree relies on.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

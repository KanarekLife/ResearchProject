package jobs

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// The jobs branch: a job file waits in pending/, is played from running/ and
// ends in archive/ or, with an error.txt, in errors/<job>/. The file keeps its
// name throughout; the job is that name without the extension.
const (
	pendingDir = "pending"
	runningDir = "running"
	archiveDir = "archive"
	errorsDir  = "errors"
	errorFile  = "error.txt"
)

var jobExts = []string{".yaml", ".yml"}

// Claim moves the first pending job by name to running/; a nil job when none
// is pending. A push lost to another worker retries on the next job.
func (b Branch) Claim(host string) (*Claimed, error) {
	var c *Claimed
	err := b.Store.Commit(func(dir string) (string, error) {
		c = nil
		names := jobNames(filepath.Join(dir, pendingDir))
		if len(names) == 0 {
			return "", nil
		}
		file, _ := findJob(filepath.Join(dir, pendingDir), names[0])
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		c = &Claimed{Name: names[0], Data: data}
		return "chore(jobs): start " + c.Name + " on " + host, move(file, filepath.Join(dir, runningDir))
	})
	return c, err
}

// Running syncs the branch and reads the jobs in running/, by name.
func (b Branch) Running() (map[string][]byte, error) {
	if err := b.Store.Sync(); err != nil {
		return nil, err
	}
	dir := filepath.Join(b.Store.Dir, runningDir)
	jobs := map[string][]byte{}
	for _, name := range jobNames(dir) {
		file, _ := findJob(dir, name)
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		jobs[name] = data
	}
	return jobs, nil
}

// Archive moves a played job from running/ to archive/.
func (b Branch) Archive(name string) error {
	return b.Store.Commit(func(dir string) (string, error) {
		file, err := findJob(filepath.Join(dir, runningDir), name)
		if err != nil {
			return "", err
		}
		return "chore(jobs): finish " + name, move(file, filepath.Join(dir, archiveDir))
	})
}

// Fail moves a job from running/ to errors/<job>/ with report as its error.txt.
func (b Branch) Fail(name, report string) error {
	return b.Store.Commit(func(dir string) (string, error) {
		file, err := findJob(filepath.Join(dir, runningDir), name)
		if err != nil {
			return "", err
		}
		errDir := filepath.Join(dir, errorsDir, name)
		if err := move(file, errDir); err != nil {
			return "", err
		}
		return "chore(jobs): " + name + " failed", os.WriteFile(filepath.Join(errDir, errorFile), []byte(report), 0o644)
	})
}

// ArchiveFailed archives a job of errors/ whose run a manual resume has now
// published. Nothing happens when the job is not in errors/.
func (b Branch) ArchiveFailed(name string) error {
	return b.Store.Commit(func(dir string) (string, error) {
		errDir := filepath.Join(dir, errorsDir, name)
		file, err := findJob(errDir, name)
		if err != nil {
			return "", nil
		}
		if err := move(file, filepath.Join(dir, archiveDir)); err != nil {
			return "", err
		}
		return "chore(jobs): finish " + name + " after a resume", os.RemoveAll(errDir)
	})
}

// jobNames are the jobs in dir, sorted by name.
func jobNames(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, e := range entries {
		ext := filepath.Ext(e.Name())
		if !e.IsDir() && slices.Contains(jobExts, ext) {
			names = append(names, strings.TrimSuffix(e.Name(), ext))
		}
	}
	return names
}

// findJob is the path of the job file of that name in dir.
func findJob(dir, name string) (string, error) {
	for _, ext := range jobExts {
		file := filepath.Join(dir, name+ext)
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}
	return "", errors.New("job " + name + " not in " + filepath.Base(dir))
}

// move moves file into dir, creating dir.
func move(file, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.Rename(file, filepath.Join(dir, filepath.Base(file)))
}

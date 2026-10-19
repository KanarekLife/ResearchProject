package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"mcbench/benchmark"
	"mcbench/trace"
	"mcbench/ui"
)

const (
	reportUsage  = "usage: mcbench report [GAMES.jsonl | DIR]...\n"
	analyzeUsage = "usage: mcbench analyze [GAMES.jsonl | DIR]...\n"
)

// cmdReport summarizes the scores of games.jsonl files.
func cmdReport(args []string) error {
	return showRecords(cmdNameReport, reportUsage, args, benchmark.Report)
}

// cmdAnalyze summarizes how the players of games.jsonl files behaved.
func cmdAnalyze(args []string) error {
	return showRecords(cmdNameAnalyze, analyzeUsage, args, benchmark.Analyze)
}

// showRecords reads games.jsonl files and writes them with show. Each
// argument is a games file or a directory searched recursively for them; with
// none, every run under the directories of -source is read.
func showRecords(name, usage string, args []string, show func(io.Writer, []benchmark.Record, ui.Theme)) error {
	flags, cfg, err := newFlagSet(name, args)
	if err != nil {
		return err
	}
	flags.Usage = func() {
		flags.Output().Write([]byte(usage))
		flags.PrintDefaults()
	}
	source := sourceFlag(flags, cfg)
	flags.Parse(args)
	paths := flags.Args()
	var files []string
	if len(paths) == 0 {
		dirs, err := resultDirs(*source, cfg.Run.Out, store(cfg))
		if err != nil {
			return err
		}
		if files, err = sourceGamesFiles(dirs); err != nil {
			return err
		}
	}
	for _, p := range paths {
		found, err := gamesFiles(p)
		if err != nil {
			return err
		}
		files = append(files, found...)
	}
	var all []benchmark.Record
	for _, f := range files {
		recs, err := benchmark.ReadRecords(f)
		if err != nil {
			return err
		}
		all = append(all, recs...)
	}
	show(os.Stdout, all, terminalTheme(os.Stdout))
	return nil
}

// sourceGamesFiles skips a later dir's copy of a run an earlier dir published.
func sourceGamesFiles(dirs []string) ([]string, error) {
	var out []string
	published := map[string]bool{}
	for _, dir := range dirs {
		files, err := gamesFiles(dir)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if !published[filepath.Base(filepath.Dir(f))] {
				out = append(out, f)
			}
		}
		for _, f := range files {
			if name := trace.Published(filepath.Dir(f)); name != "" {
				published[name] = true
			}
		}
	}
	return out, nil
}

// gamesFiles returns path itself when it is a file, else every games.jsonl
// below it.
func gamesFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == gamesFile {
			files = append(files, p)
		}
		return err
	})
	return files, err
}

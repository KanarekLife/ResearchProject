package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"mcbench/benchmark"
)

const reportUsage = "usage: mcbench report [GAMES.jsonl | DIR]...\n"

// cmdReport summarizes games.jsonl files. Each argument is a games file or a
// directory searched recursively for them; with none, every run under the
// results directory is reported.
func cmdReport(args []string) error {
	flags, cfg, err := newFlagSet(cmdNameReport, args)
	if err != nil {
		return err
	}
	flags.Usage = func() {
		flags.Output().Write([]byte(reportUsage))
		flags.PrintDefaults()
	}
	flags.Parse(args)
	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{cfg.Run.Out}
	}
	var all []benchmark.Record
	for _, p := range paths {
		files, err := gamesFiles(p)
		if err != nil {
			return err
		}
		for _, f := range files {
			recs, err := benchmark.ReadRecords(f)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}
	}
	benchmark.Report(os.Stdout, all, terminalTheme())
	return nil
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

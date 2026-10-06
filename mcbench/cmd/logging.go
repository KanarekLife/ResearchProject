package main

import (
	"log/slog"
	"os"
)

const (
	defaultLogLevel = "info"
	logLevelUsage   = "log level on stderr: debug (model requests), info (game actions and choices), warn, error"
)

// setupLogging installs the default structured logger. Every package logs
// through slog.Default, so the level is set in one place.
func setupLogging(level string) error {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
	return nil
}

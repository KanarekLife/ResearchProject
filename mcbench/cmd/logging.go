package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"mcbench/ui"
)

const (
	defaultLogLevel = "info"
	logLevelUsage   = "log level on stderr: debug (model requests), info (game actions and choices), warn, error"
	// messageWidth is where the key=value column starts, so pairs line up.
	// Longer messages are not cut; their pairs follow on the same line.
	messageWidth = 60
)

// setupLogging installs the default structured logger. Every package logs
// through slog.Default, so the level is set in one place.
func setupLogging(level string) error {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	slog.SetDefault(slog.New(&consoleHandler{mu: &sync.Mutex{}, w: os.Stderr, level: l, theme: terminalTheme(os.Stderr)}))
	return nil
}

// consoleHandler writes one line per record for a person reading a terminal:
// time, level, message, then key=value pairs in their own column. Groups are
// not used by mcbench, so they are ignored.
type consoleHandler struct {
	mu    *sync.Mutex // shared by all clones, so lines from parallel games do not interleave
	w     io.Writer
	level slog.Level
	theme ui.Theme
	attrs []string // pairs added by With, written before each record's own pairs
}

func (h *consoleHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	pairs := slices.Clone(h.attrs)
	r.Attrs(func(a slog.Attr) bool {
		pairs = appendPair(pairs, a, h.theme)
		return true
	})
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s %s", h.theme.Dim(r.Time.Format(time.TimeOnly)), h.levelText(r.Level), r.Message)
	if len(pairs) > 0 {
		fmt.Fprintf(&b, "%*s%s", max(0, messageWidth-len(r.Message)), "", " "+strings.Join(pairs, " "))
	}
	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := *h
	c.attrs = slices.Clone(h.attrs)
	for _, a := range attrs {
		c.attrs = appendPair(c.attrs, a, h.theme)
	}
	return &c
}

func (h *consoleHandler) WithGroup(string) slog.Handler { return h }

// levelText pads before painting: escape codes would otherwise skew the column.
func (h *consoleHandler) levelText(l slog.Level) string {
	s := fmt.Sprintf("%-5s", l)
	switch {
	case l >= slog.LevelError:
		return h.theme.Red(s)
	case l >= slog.LevelWarn:
		return h.theme.Yellow(s)
	case l >= slog.LevelInfo:
		return h.theme.Cyan(s)
	default:
		return h.theme.Dim(s)
	}
}

// appendPair adds "key=value" to pairs, quoting values that would be hard to
// read or split apart: empty, containing spaces, quotes or '='.
func appendPair(pairs []string, a slog.Attr, theme ui.Theme) []string {
	if a.Equal(slog.Attr{}) {
		return pairs
	}
	v := a.Value.Resolve().String()
	if v == "" || strings.ContainsAny(v, " \t\n\"=") {
		v = strconv.Quote(v)
	}
	return append(pairs, theme.Gray(a.Key+"=")+v)
}

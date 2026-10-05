package main

import (
	"os"
	"strings"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// flagValue reads -name value or -name=value from raw args before flags are
// defined, so config can supply the flag defaults.
func flagValue(args []string, name, def string) string {
	for i, a := range args {
		if a == "-"+name || a == "--"+name {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if v, ok := strings.CutPrefix(a, "-"+name+"="); ok {
			return v
		}
		if v, ok := strings.CutPrefix(a, "--"+name+"="); ok {
			return v
		}
	}
	return def
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	return strings.Join(lines[:min(n, len(lines))], "\n  ")
}

// sanitize makes a player name safe for a directory name.
func sanitize(s string) string {
	return strings.NewReplacer("/", "_", ":", "_", " ", "_").Replace(s)
}

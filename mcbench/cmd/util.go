package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"mcbench/ui"
)

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

// terminalTheme colours output only when f is a terminal and NO_COLOR is
// unset, so piped output stays plain.
func terminalTheme(f *os.File) ui.Theme {
	fi, err := f.Stat()
	tty := err == nil && fi.Mode()&os.ModeCharDevice != 0
	return ui.Theme{Color: tty && os.Getenv("NO_COLOR") == ""}
}

// parseSeeds reads a comma-separated list of seeds; empty means none.
func parseSeeds(s string) ([]uint64, error) {
	if s == "" {
		return nil, nil
	}
	var out []uint64
	for _, f := range strings.Split(s, ",") {
		n, err := strconv.ParseUint(strings.TrimSpace(f), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("-pick: %q is not a seed", f)
		}
		out = append(out, n)
	}
	return out, nil
}

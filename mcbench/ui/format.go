package ui

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Bar draws cur/total as a filled bar of the given cell width.
func Bar(cur, total, cells int) string {
	if total <= 0 {
		return strings.Repeat("░", cells)
	}
	cur = min(max(cur, 0), total)
	filled := (cur*cells + total - 1) / total
	return strings.Repeat("█", filled) + strings.Repeat("░", cells-filled)
}

// Meter draws a fraction in [0, 1] as a bar of the given cell width.
func Meter(frac float64, cells int) string {
	return Bar(int(math.Round(frac*100)), 100, cells)
}

// Thousands writes a non-negative count with a comma between every three
// digits: 2,705,233.
func Thousands(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + thousandsSep + s[i:]
	}
	return s
}

const thousandsSep = ","

// Elapsed renders a duration to the second, e.g. 4m6s.
func Elapsed(d time.Duration) string { return d.Round(time.Second).String() }

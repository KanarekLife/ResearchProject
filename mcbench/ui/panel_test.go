package ui

import "testing"

// Every row of a box must span the same columns, whether or not it carries
// colour or had to be wrapped.
func TestBoxRowsAlign(t *testing.T) {
	for _, theme := range []Theme{{Color: false}, {Color: true}} {
		lines := []string{
			theme.Bold("short"),
			theme.Dim("a long dimmed line that has to be wrapped onto several rows inside the box frame"),
			"· plain ünïcode",
		}
		for i, row := range theme.Box("Title", lines, 40) {
			if got := visibleLen(row); got != 40 {
				t.Errorf("color=%v row %d is %d columns, want 40: %q", theme.Color, i, got, row)
			}
		}
	}
}

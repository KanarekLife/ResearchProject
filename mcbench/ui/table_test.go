package ui

import "testing"

// Table columns must line up by visible width, so coloured cells do not shift
// the columns after them.
func TestTableAlignsStyledCells(t *testing.T) {
	theme := Theme{Color: true}
	rows := theme.Table([]Column{{Title: "result"}, {Title: "n", Right: true}}, [][]string{{theme.Green("won"), "1"}, {"agent_error", "2"}})
	if a, b := visibleLen(rows[1]), visibleLen(rows[2]); a != b {
		t.Errorf("rows are %d and %d columns wide: %q", a, b, rows)
	}
}

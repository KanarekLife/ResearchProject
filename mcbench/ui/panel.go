package ui

import (
	"fmt"
	"io"
	"strings"
)

// Box frames lines under a title. Lines wider than the box are wrapped on word
// boundaries; ANSI styling is carried onto the continuation rows. A width of 0
// fits the box to its widest line.
func (t Theme) Box(title string, lines []string, width int) []string {
	if width == 0 {
		for _, l := range lines {
			width = max(width, visibleLen(l)+4)
		}
		width = max(width, visibleLen(title)+5)
	}
	inner := width - 4
	head := " " + title + " "
	top := "┌─" + t.Bold(head) + strings.Repeat("─", max(width-3-visibleLen(head), 0)) + "┐"
	out := []string{top}
	end := ""
	if t.Color {
		end = reset
	}
	for _, l := range lines {
		for _, row := range wrap(l, inner) {
			pad := strings.Repeat(" ", max(inner-visibleLen(row), 0))
			out = append(out, "│ "+row+end+pad+" │")
		}
	}
	return append(out, "└"+strings.Repeat("─", width-2)+"┘")
}

// Side places two panels next to each other, padding the shorter one.
func Side(left, right []string, gap int) []string {
	w := 0
	for _, l := range left {
		w = max(w, visibleLen(l))
	}
	var out []string
	for i := range max(len(left), len(right)) {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		out = append(out, l+strings.Repeat(" ", w-visibleLen(l)+gap)+r)
	}
	return out
}

// Print writes panels one after another.
func Print(w io.Writer, panels ...[]string) {
	for _, p := range panels {
		fmt.Fprintln(w, strings.Join(p, "\n"))
	}
}

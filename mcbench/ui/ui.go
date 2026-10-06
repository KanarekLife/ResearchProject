// Package ui draws terminal panels: boxed sections, side-by-side columns,
// bars and ANSI colour. It knows nothing about the game.
package ui

import (
	"strings"
	"unicode/utf8"
)

// Width is the full width of a screen of panels, in columns.
const Width = 100

const (
	esc   = "\x1b["
	reset = esc + "0m"
)

// Theme paints text; with Color off every style returns its text unchanged, so
// output stays clean when piped.
type Theme struct{ Color bool }

func (t Theme) paint(code, s string) string {
	if !t.Color || s == "" {
		return s
	}
	return esc + code + "m" + s + reset
}

func (t Theme) Bold(s string) string   { return t.paint("1", s) }
func (t Theme) Dim(s string) string    { return t.paint("2", s) }
func (t Theme) Red(s string) string    { return t.paint("31", s) }
func (t Theme) Green(s string) string  { return t.paint("32", s) }
func (t Theme) Yellow(s string) string { return t.paint("33", s) }
func (t Theme) Cyan(s string) string   { return t.paint("36", s) }

// Bar draws cur/total as a filled bar of the given cell width.
func Bar(cur, total, cells int) string {
	if total <= 0 {
		return strings.Repeat("░", cells)
	}
	cur = min(max(cur, 0), total)
	filled := (cur*cells + total - 1) / total
	return strings.Repeat("█", filled) + strings.Repeat("░", cells-filled)
}

// Box frames lines under a title. Lines wider than the box are wrapped on word
// boundaries; ANSI styling is carried onto the continuation rows.
func (t Theme) Box(title string, lines []string, width int) []string {
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

// Wrap breaks plain text into lines of at most width columns.
func Wrap(text string, width int) []string { return wrap(text, width) }

// wrap breaks s on spaces into rows of at most width visible columns. A row
// that continues a styled run re-opens the style it was in.
func wrap(s string, width int) []string {
	var rows []string
	var row strings.Builder
	rowLen, style := 0, ""
	for _, word := range strings.Split(s, " ") {
		wl := visibleLen(word)
		if rowLen > 0 && rowLen+1+wl > width {
			rows = append(rows, row.String())
			row.Reset()
			row.WriteString(style)
			rowLen = 0
		}
		if rowLen > 0 {
			row.WriteByte(' ')
			rowLen++
		}
		row.WriteString(word)
		rowLen += wl
		style = trackStyle(style, word)
	}
	return append(rows, row.String())
}

// trackStyle returns the SGR sequence in force after s, given the one before.
func trackStyle(style, s string) string {
	for {
		i := strings.Index(s, esc)
		if i < 0 {
			return style
		}
		end := strings.IndexByte(s[i:], 'm')
		if end < 0 {
			return style
		}
		seq := s[i : i+end+1]
		if seq == reset {
			style = ""
		} else {
			style += seq
		}
		s = s[i+end+1:]
	}
}

// visibleLen counts the columns s occupies, ignoring ANSI sequences.
func visibleLen(s string) int {
	n := 0
	for len(s) > 0 {
		if strings.HasPrefix(s, esc) {
			if end := strings.IndexByte(s, 'm'); end >= 0 {
				s = s[end+1:]
				continue
			}
		}
		_, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		n++
	}
	return n
}

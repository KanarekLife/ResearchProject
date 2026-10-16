package ui

import (
	"strings"
	"unicode/utf8"
)

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

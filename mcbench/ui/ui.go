// Package ui draws terminal panels: boxed sections, side-by-side columns,
// tables, bars, number formats and ANSI colour. It knows nothing about the
// game.
package ui

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

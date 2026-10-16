package ui

import "strings"

// columnGap is the space between table columns.
const columnGap = 2

// Column is a table column: its header and whether its cells are
// right-aligned (for numbers).
type Column struct {
	Title string
	Right bool
}

// Cols makes left-aligned columns from titles.
func Cols(titles ...string) []Column {
	cols := make([]Column, len(titles))
	for i, title := range titles {
		cols[i] = Column{Title: title}
	}
	return cols
}

// Table aligns a bold header and rows into columns. Cells may carry ANSI
// styling; widths are measured in visible columns.
func (t Theme) Table(cols []Column, rows [][]string) []string {
	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = t.Bold(c.Title)
	}
	rows = append([][]string{header}, rows...)
	widths := columnWidths(rows)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		var b strings.Builder
		for i, cell := range row {
			pad := strings.Repeat(" ", widths[i]-visibleLen(cell))
			switch {
			case i < len(cols) && cols[i].Right:
				b.WriteString(pad + cell)
			case i < len(row)-1:
				b.WriteString(cell + pad)
			default:
				b.WriteString(cell)
			}
			if i < len(row)-1 {
				b.WriteString(strings.Repeat(" ", columnGap))
			}
		}
		out = append(out, b.String())
	}
	return out
}

func columnWidths(rows [][]string) []int {
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if i == len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], visibleLen(cell))
		}
	}
	return widths
}

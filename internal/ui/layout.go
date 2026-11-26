package ui

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// layout controls how rows are drawn. The responsive screen lowers the
// detail step by step when the card does not fit.
type layout struct {
	// icons adds the icon column.
	icons bool
}

// labelGap separates the label column from the value.
const labelGap = "  "

// labelWidth returns the width of the label column: the widest label of
// all rows of all sections, so values line up across the card.
func labelWidth(ctx beautify.Context, sections [][]row) int {
	f := ctx.Factory()
	width := 0
	for _, rows := range sections {
		for _, r := range rows {
			width = max(width, f.Text(oneLine(r.label)).Width())
		}
	}
	return width
}

// renderRows draws the rows of one section, one line per row.
func renderRows(ctx beautify.Context, th theme, lay layout, rows []row, labels int) []dra.Model {
	f := ctx.Factory()
	lines := make([]dra.Model, 0, len(rows))
	for _, r := range rows {
		var parts []dra.Model
		if lay.icons {
			icon := "  "
			if r.icon != 0 {
				icon = string(r.icon) + " "
			}
			parts = append(parts, f.Styled(icon, th.icon))
		}
		label := dra.Pad(f.Styled(oneLine(r.label), th.label), labels, dra.AlignLeft)
		parts = append(parts, label, f.Text(labelGap))
		for _, s := range r.value {
			parts = append(parts, f.Styled(oneLine(s.text), th.style(s.role)))
		}
		lines = append(lines, dra.Concat(parts...))
	}
	return lines
}

// style returns the style of a role.
func (th theme) style(r role) dra.Style {
	switch r {
	case roleMuted:
		return th.muted
	case roleGood:
		return th.good
	case roleWarning:
		return th.warning
	case roleDanger:
		return th.danger
	}
	return th.value
}

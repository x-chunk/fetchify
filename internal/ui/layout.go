package ui

import (
	"fmt"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/v2/internal/config"
)

// layout controls how rows are drawn. The responsive screen lowers the
// detail step by step when the card does not fit.
type layout struct {
	// icons adds the icon column.
	icons bool
	// details shows the segments marked as details.
	details bool
	// bars draws the bar of meters; without it only the numbers remain.
	bars bool
	// barWidth is the width of bars in cells.
	barWidth int
	// barStyle selects the characters of bars.
	barStyle config.BarStyle
	// thresholds are the percentages where a fill turns to warning and
	// danger.
	thresholds [2]int
	// truncate cuts rows that are wider than the bound and marks the cut
	// with an ellipsis.
	truncate bool
}

// meterColumns are the widths of the right-aligned columns of the meters
// of one section.
type meterColumns struct {
	used, total, percent int
}

// meterText holds the formatted numbers of a meter.
type meterText struct {
	used, total, unit, percent string
}

// formatMeter formats the numbers of m. used and total are empty for a
// meter without sizes.
func formatMeter(m meter) meterText {
	t := meterText{percent: fmt.Sprintf("%d%%", m.percent)}
	if m.sizes {
		unit, div := sizeUnit(m.total)
		t.used = formatAmount(float64(m.used) / div)
		t.total = formatAmount(float64(m.total) / div)
		t.unit = unit
	}
	return t
}

// columnsOf returns the column widths of the meters of rows.
func columnsOf(rows []row) meterColumns {
	var c meterColumns
	for _, r := range rows {
		if r.meter == nil {
			continue
		}
		t := formatMeter(*r.meter)
		c.used = max(c.used, len(t.used))
		c.total = max(c.total, len(t.total))
		c.percent = max(c.percent, len(t.percent))
	}
	return c
}

// renderMeter draws "bar  used / total unit  percent" with the numbers
// right-aligned in the columns of the section.
func renderMeter(ctx beautify.Context, th theme, lay layout, m meter, cols meterColumns) dra.Model {
	f := ctx.Factory()
	t := formatMeter(m)
	var parts []dra.Model
	if lay.bars {
		fill := th.style(level(m, lay.thresholds))
		b, _ := bar{percent: m.percent, width: lay.barWidth, style: lay.barStyle, fill: fill, empty: th.muted}.Render(ctx.Unbounded())
		parts = append(parts, b, f.Text("  "))
	}
	if m.sizes {
		parts = append(parts,
			dra.Pad(f.Text(t.used), cols.used, dra.AlignRight),
			f.Text(" / "),
			dra.Pad(f.Text(t.total), cols.total, dra.AlignRight),
			f.Text(" "+t.unit+"  "))
	}
	parts = append(parts, dra.Pad(f.Text(t.percent), cols.percent, dra.AlignRight))
	return dra.Concat(parts...)
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
	cols := columnsOf(rows)
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
		value := r.value
		if !lay.details {
			value = withoutDetails(value)
		}
		if r.meter != nil {
			parts = append(parts, renderMeter(ctx, th, lay, *r.meter, cols))
			if len(value) > 0 {
				parts = append(parts, f.Text("  "))
			}
		}
		for _, s := range value {
			parts = append(parts, f.Styled(oneLine(s.text), th.style(s.role)))
		}
		if r.badge != nil {
			if r.meter != nil || len(value) > 0 {
				parts = append(parts, f.Text(" "))
			}
			b, err := blocks.Badge(r.badge.text, th.badgeStyle(r.badge.look)).Render(ctx.Unbounded())
			if err == nil {
				parts = append(parts, b)
			}
		}
		lines = append(lines, dra.Concat(parts...))
	}
	return lines
}

// withoutDetails drops the detail segments of a value. The separator
// of a list such as "1842 (dpkg), 12 (flatpak)" stays, giving "1842, 12".
func withoutDetails(value []segment) []segment {
	out := make([]segment, 0, len(value))
	for _, s := range value {
		if !s.detail {
			out = append(out, s)
		}
	}
	return out
}

// badgeStyle returns the style of a badge look.
func (th theme) badgeStyle(look badgeLook) dra.Style {
	switch look {
	case badgeGood:
		return th.good.WithAttrs(dra.Reverse)
	case badgeWarning:
		return th.warning.WithAttrs(dra.Reverse)
	}
	return th.badge
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

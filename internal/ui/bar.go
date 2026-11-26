package ui

import (
	"fmt"
	"math"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/v2/internal/config"
)

// Fixed thresholds of an inverted meter such as a battery: below the
// first the level is danger, below the second warning.
var invertedThresholds = [2]int{20, 40}

// bar is a usage bar of a fixed width. With less room than its width it
// shrinks to the bound.
type bar struct {
	percent int
	width   int
	style   config.BarStyle
	fill    dra.Style
	empty   dra.Style
}

// Render implements beautify.Composable.
func (b bar) Render(ctx beautify.Context) (dra.Model, error) {
	width := b.width
	if bound, ok := ctx.Width(); ok {
		width = min(width, bound)
	}
	if width <= 0 {
		return dra.Model{}, nil
	}
	fillChar, emptyChar := "█", "░"
	if b.style == config.BarThin {
		fillChar, emptyChar = "━", "━"
	}
	f := ctx.Factory()
	filled := barCells(b.percent, width)
	var parts []dra.Model
	if filled > 0 {
		parts = append(parts, dra.Repeat(f.Styled(fillChar, b.fill), filled))
	}
	if filled < width {
		parts = append(parts, dra.Repeat(f.Styled(emptyChar, b.empty), width-filled))
	}
	return dra.Concat(parts...), nil
}

// barCells returns the number of filled cells of a bar. A share above 0
// fills at least one cell and a share below 100 leaves at least one
// empty, so "almost empty" and "almost full" stay visible.
func barCells(percent, width int) int {
	if width <= 0 {
		return 0
	}
	n := int(math.Round(float64(width) * float64(percent) / 100))
	if percent > 0 {
		n = max(n, 1)
	}
	if percent < 100 {
		n = min(n, width-1)
	}
	return min(max(n, 0), width)
}

// level returns the role of a meter's fill: good, warning or danger.
func level(m meter, thresholds [2]int) role {
	if m.inverted {
		switch {
		case m.percent < invertedThresholds[0]:
			return roleDanger
		case m.percent < invertedThresholds[1]:
			return roleWarning
		}
		return roleGood
	}
	switch {
	case m.percent < thresholds[0]:
		return roleGood
	case m.percent < thresholds[1]:
		return roleWarning
	}
	return roleDanger
}

// Binary units of sizes.
const (
	bytesMiB = 1 << 20
	bytesGiB = 1 << 30
	bytesTiB = 1 << 40
)

// sizeUnit chooses the unit of a meter by its total: MiB below 1 GiB,
// TiB from 1024 GiB, GiB otherwise.
func sizeUnit(total uint64) (string, float64) {
	switch {
	case total < bytesGiB:
		return "MiB", bytesMiB
	case total >= bytesTiB:
		return "TiB", bytesTiB
	}
	return "GiB", bytesGiB
}

// formatAmount shows one decimal below 100 and none from 100.
func formatAmount(v float64) string {
	if v < 99.95 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.0f", v)
}

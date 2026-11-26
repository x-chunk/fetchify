package ui

import (
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
)

// The styles of the screen, named by role so that the look is defined in
// one place.
var (
	styleKey      = dra.Foreground(dra.Yellow)
	styleUser     = dra.Foreground(dra.Red)
	styleAt       = dra.Foreground(dra.Green)
	styleHost     = dra.Foreground(dra.Cyan)
	styleRule     = dra.Foreground(dra.Green)
	styleWarning  = dra.Foreground(dra.Red)
	styleMissing  = dra.Foreground(dra.Red)
	styleUsed     = dra.Foreground(dra.Green)
	styleTotal    = dra.Foreground(dra.Cyan)
	styleCategory = dra.Foreground(dra.Magenta)
)

// paletteColors is the order of the color swatches under the info list.
var paletteColors = []dra.Color{
	dra.Black, dra.Red, dra.Green, dra.Yellow,
	dra.Blue, dra.Magenta, dra.Cyan, dra.White,
}

// theme holds the styles of the card by role. It is resolved once per
// run from the configuration and the distribution; no part of the screen
// uses a color directly.
//
// Good, Warning, Danger, Muted and the palette use the 16-color palette
// so that they follow the terminal theme. The accent is 24-bit: it
// carries the identity of the distribution, and beautify converts it for
// weaker terminals.
type theme struct {
	accent  dra.Color
	frame   dra.Style
	title   dra.Style
	label   dra.Style
	icon    dra.Style
	value   dra.Style
	muted   dra.Style
	caption dra.Style
	good    dra.Style
	warning dra.Style
	danger  dra.Style
	// badge is the neutral look of a badge.
	badge dra.Style
	// gradient colors the frame instead of frame when it has stops.
	gradient dra.Gradient
}

// newTheme resolves the theme. The accent is the configured color or, for
// "auto" and anything that does not parse, the accent of d. root turns
// the caption red, which replaces the old warning text.
func newTheme(cfg config.Config, d distro.Distro, root bool) theme {
	accent := d.Accent
	if cfg.Colors.Accent != config.AccentAuto {
		if c, err := dra.Hex(cfg.Colors.Accent); err == nil {
			accent = c
		}
	}
	muted := dra.Foreground(dra.BrightBlack)
	caption := dra.Style{Attrs: dra.Bold}
	if root {
		caption = dra.Foreground(dra.Red).WithAttrs(dra.Bold)
	}
	frame := dra.Foreground(accent)
	var gradient dra.Gradient
	switch stops := frameColors(cfg, d.Colors); len(stops) {
	case 0:
	case 1:
		frame = dra.Foreground(stops[0])
	default:
		gradient = dra.NewGradient(axes[cfg.Border.Direction], stops...)
	}
	return theme{
		accent:   accent,
		frame:    frame,
		gradient: gradient,
		title:    dra.Foreground(accent).WithAttrs(dra.Bold),
		label:    dra.Foreground(accent),
		icon:     dra.Foreground(accent),
		muted:    muted,
		caption:  caption,
		good:     dra.Foreground(dra.Green),
		warning:  dra.Foreground(dra.Yellow),
		danger:   dra.Foreground(dra.Red),
		badge:    muted.WithAttrs(dra.Reverse),
	}
}

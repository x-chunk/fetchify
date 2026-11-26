package ui

import (
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
)

// rainbow are the stops of border.color = "rainbow".
var rainbow = []string{"#ff5f5f", "#ffbf5f", "#ffff5f", "#5fff5f", "#5fffff", "#5f5fff", "#ff5fff"}

// axes maps the configured directions to gradient axes.
var axes = map[config.Direction]dra.Axis{
	config.Vertical:   dra.AxisVertical,
	config.Horizontal: dra.AxisHorizontal,
	config.Diagonal:   dra.AxisDiagonal,
}

// hexColors parses colors that config has validated. Anything that does
// not parse is skipped.
func hexColors(list []string) []dra.Color {
	colors := make([]dra.Color, 0, len(list))
	for _, s := range list {
		if c, err := dra.Hex(s); err == nil {
			colors = append(colors, c)
		}
	}
	return colors
}

// solidColors returns the colors that a gradient can use: TerminalDefault
// has no RGB value and is left out.
func solidColors(list []dra.Color) []dra.Color {
	out := make([]dra.Color, 0, len(list))
	for _, c := range list {
		if c.IsSet() && c != dra.TerminalDefault {
			out = append(out, c)
		}
	}
	return out
}

// frameColors returns the stops of the frame for the configuration. One
// stop means a solid frame; none means the accent.
func frameColors(cfg config.Config, logo []dra.Color) []dra.Color {
	switch cfg.Border.Color {
	case config.BorderLogo:
		return solidColors(logo)
	case config.BorderRainbow:
		return hexColors(rainbow)
	case config.BorderCustom:
		return hexColors(cfg.Border.Colors)
	}
	return nil
}

// recolorLogo layers a gradient over a logo model. Spaces get the color
// too, which is invisible because no background is set.
func recolorLogo(m dra.Model, g dra.Gradient) dra.Model {
	if g.IsZero() {
		return m
	}
	w, h := m.Width(), m.Height()
	return dra.RestyleFunc(m, func(x, y int) dra.Style {
		return dra.Foreground(g.At(x, y, w, h))
	})
}

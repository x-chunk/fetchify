package ui

import "go.xchunk.org/beautify/dra"

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

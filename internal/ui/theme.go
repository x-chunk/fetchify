package ui

import "github.com/x-chunk/fetchify/internal/render"

// The styles of the screen, named by role so that the look is defined in
// one place.
var (
	styleKey      = render.Foreground(render.Yellow)
	styleUser     = render.Foreground(render.Red)
	styleAt       = render.Foreground(render.Green)
	styleHost     = render.Foreground(render.Cyan)
	styleRule     = render.Foreground(render.Green)
	styleWarning  = render.Foreground(render.Red)
	styleMissing  = render.Foreground(render.Red)
	styleUsed     = render.Foreground(render.Green)
	styleTotal    = render.Foreground(render.Cyan)
	styleCategory = render.Foreground(render.Magenta)
)

// paletteColors is the order of the color swatches under the info list.
var paletteColors = []render.Color{
	render.Black, render.Red, render.Green, render.Yellow,
	render.Blue, render.Magenta, render.Cyan, render.White,
}

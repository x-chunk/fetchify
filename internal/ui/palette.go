package ui

import (
	"strings"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"
)

// paletteSize is the number of terminal colors shown: the normal colors
// 0-7. They are always indexed, never 24-bit, because the palette exists
// to show the terminal's own colors.
const paletteSize = 8

// paletteDots returns eight dots separated by spaces, the footer of the
// card in the dots style.
func paletteDots(f *dra.Factory) dra.Model {
	parts := make([]dra.Model, 0, 2*paletteSize-1)
	for i := range paletteSize {
		if i > 0 {
			parts = append(parts, f.Text(" "))
		}
		parts = append(parts, f.Styled("●", dra.Foreground(dra.ANSI(uint8(i)))))
	}
	return dra.Concat(parts...)
}

// paletteBlocks returns the blocks style: two identical rows of eight
// swatches under the card, indented by one cell.
func paletteBlocks() beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		f := ctx.Factory()
		swatches := make([]dra.Model, paletteSize)
		for i := range swatches {
			swatches[i] = f.Styled(strings.Repeat("█", 3), dra.Foreground(dra.ANSI(uint8(i))))
		}
		row := dra.Concat(append([]dra.Model{f.Text(" ")}, swatches...)...)
		return blocks.Model(dra.Stack(dra.AlignLeft, row, row)).Render(ctx)
	})
}

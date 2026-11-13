package render

import "strings"

// Stack places blocks under each other, in order. Empty blocks contribute
// nothing. Lines are not padded.
func Stack(blocks ...Block) Block {
	var out Block
	for _, block := range blocks {
		out = append(out, block...)
	}
	return out
}

// Join places blocks side by side, top-aligned, with gap blank cells
// between neighbors. A negative gap is treated as 0. Empty blocks are
// skipped together with their gap.
//
// The result is as tall as the tallest block. Every line of every block
// except the last one is padded on the right to that block's width, so
// the columns line up; lines of the last block are not padded. Padding is
// unstyled.
func Join(gap int, blocks ...Block) Block {
	gap = max(gap, 0)

	var columns []Block
	height := 0
	for _, block := range blocks {
		if block.Height() == 0 {
			continue
		}
		columns = append(columns, block)
		height = max(height, block.Height())
	}

	out := make(Block, height)
	for i, column := range columns {
		last := i == len(columns)-1
		width := column.Width()
		for row := range out {
			var line Line
			if row < len(column) {
				line = column[row]
			}
			out[row] = append(out[row], line...)
			if last {
				continue
			}
			if pad := width - line.Width() + gap; pad > 0 {
				out[row] = append(out[row], Text(strings.Repeat(" ", pad)))
			}
		}
	}
	return out
}

// Repeat returns a span that repeats unit until it is exactly width cells
// wide, drawn with st. The last copy of unit is cut when width is not a
// multiple of the unit width. It returns an empty span when width is not
// positive or unit is empty.
func Repeat(unit string, width int, st Style) Span {
	runes := []rune(sanitize(unit))
	if width <= 0 || len(runes) == 0 {
		return Span{style: st}
	}
	var sb strings.Builder
	for i := range width {
		sb.WriteRune(runes[i%len(runes)])
	}
	return Span{text: sb.String(), style: st}
}

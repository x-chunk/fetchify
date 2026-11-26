package ui

import (
	"fmt"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
)

// footerIndent is the indent of the footer under the card.
const footerIndent = 2

// footer returns the line under the card with the collection time,
// rounded to milliseconds.
func footer(th theme, elapsed time.Duration) beautify.Composable {
	text := fmt.Sprintf("collected in %d ms", elapsed.Round(time.Millisecond).Milliseconds())
	return blocks.Padding(
		blocks.Text(text, blocks.TextStyle(th.muted), blocks.TextOverflow(blocks.OverflowTruncate)),
		blocks.PaddingSides(0, 0, 0, footerIndent),
	)
}

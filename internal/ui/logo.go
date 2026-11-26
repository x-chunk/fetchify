package ui

import (
	"strings"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
)

// logoLines is the ASCII-art koala shown to the left of the info column.
var logoLines = []string{
	"==== Fetchify ====",
	"   __         __  ",
	" /\"  \"\\     /\"  \"\\",
	"(  (\\  )___(  /)  )",
	" \\               / ",
	" /               \\ ",
	"/    () ___ ()    \\",
	"|      (   )      |",
	" \\      \\_/      / ",
	"   \\...__!__.../   ",
	"        \" mh       ",
	"   v1.0.4-generic   ",
}

// logo returns the logo as unstyled text. Wrapping would destroy the
// drawing, so lines that do not fit are cut instead.
func logo() beautify.Composable {
	return blocks.Text(strings.Join(logoLines, "\n"), blocks.TextOverflow(blocks.OverflowKeep))
}

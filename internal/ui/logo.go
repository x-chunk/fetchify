package ui

import "github.com/x-chunk/fetchify/internal/render"

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

// logo returns the logo as an unstyled block.
func logo() render.Block {
	block := make(render.Block, len(logoLines))
	for i, line := range logoLines {
		block[i] = render.Line{render.Text(line)}
	}
	return block
}

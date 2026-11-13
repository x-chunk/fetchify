package ui

import (
	"strings"

	"github.com/x-chunk/fetchify/internal/render"
	"github.com/x-chunk/fetchify/internal/sysinfo"
)

const (
	// logoGap is the number of blank cells between the logo and the info
	// column.
	logoGap = 1
	// swatch is the text of one palette cell.
	swatch = "███"
	// paletteRows is the number of identical rows the palette is drawn
	// with, which makes the swatches roughly square.
	paletteRows = 2
)

// Build returns the whole fetch screen for info: the logo on the left and,
// next to it, the header, the info list and the color palette. The result
// ends with an empty line that separates it from the shell prompt.
func Build(info sysinfo.Info) render.Block {
	column := render.Stack(
		header(info),
		keyValue(entries(info)),
		render.Block{nil},
		palette(),
		render.Block{nil},
	)
	return render.Join(logoGap, logo(), column)
}

// header returns the "user@host" title and the rule under it. For root it
// adds a warning to the title and draws the rule in the warning style.
func header(info sysinfo.Info) render.Block {
	title := render.Line{
		render.Styled(info.User, styleUser),
		render.Styled("@", styleAt),
		render.Styled(info.Hostname, styleHost),
	}
	ruleStyle := styleRule
	if info.Root {
		title = append(title,
			render.Text(" - "),
			render.Styled("Don't login as root ^-^", styleWarning),
		)
		ruleStyle = styleWarning
	}
	rule := render.Line{render.Repeat("─", title.Width(), ruleStyle)}
	return render.Block{title, rule}
}

// palette returns the rows of color swatches.
func palette() render.Block {
	row := make(render.Line, len(paletteColors))
	for i, color := range paletteColors {
		row[i] = render.Styled(swatch, render.Foreground(color))
	}
	block := make(render.Block, paletteRows)
	for i := range block {
		block[i] = row
	}
	return block
}

// entry is one row of the info list.
type entry struct {
	key   string
	value render.Line
}

// keyValue lays entries out one per line as "key: value", with the key in
// the key style. Keys are not padded to a common width.
func keyValue(entries []entry) render.Block {
	block := make(render.Block, len(entries))
	for i, e := range entries {
		line := render.Line{render.Styled(e.key, styleKey), render.Text(": ")}
		block[i] = append(line, e.value...)
	}
	return block
}

// section produces the entries of one topic of the info list. A section
// may return no entries when it has nothing to show.
type section func(sysinfo.Info) []entry

// sections is the content of the info list, in display order. To add,
// remove or reorder what fetchify shows, edit this slice.
var sections = []section{
	osSection,
	kernelSection,
	uptimeSection,
	shellSection,
	desktopSection,
	cpuSection,
	gpuSection,
	memorySection,
	swapSection,
	diskSection,
}

// entries runs every section and returns their entries in order.
func entries(info sysinfo.Info) []entry {
	var result []entry
	for _, s := range sections {
		result = append(result, s(info)...)
	}
	return result
}

// one returns a single-entry list, the result of most sections.
func one(key string, value ...render.Span) []entry {
	return []entry{{key: key, value: value}}
}

// textOrMissing returns s as plain text, or the "Unknown" marker when s
// is blank.
func textOrMissing(s string) render.Span {
	if strings.TrimSpace(s) == "" {
		return missing("Unknown")
	}
	return render.Text(s)
}

// missing returns reason in the style that marks absent information.
func missing(reason string) render.Span {
	return render.Styled(reason, styleMissing)
}

package ui

import (
	"strings"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/sysinfo"
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
func Build(info sysinfo.Info) beautify.Template {
	column := blocks.Column([]beautify.Composable{
		header(info),
		keyValue(info),
		blocks.Space(0, 1),
		palette(),
		blocks.Space(0, 1),
	})
	return beautify.NewTemplate(blocks.Row(
		[]blocks.Cell{blocks.Auto(logo()), blocks.Auto(column)},
		blocks.RowGap(logoGap),
	))
}

// header returns the "user@host" title and the rule under it. For root it
// adds a warning to the title and draws the rule in the warning style.
// The rule is exactly as wide as the title.
func header(info sysinfo.Info) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		f := ctx.Factory()
		title := dra.Concat(
			f.Styled(oneLine(info.User), styleUser),
			f.Styled("@", styleAt),
			f.Styled(oneLine(info.Hostname), styleHost),
		)
		ruleStyle := styleRule
		if info.Root {
			title = dra.Concat(title,
				f.Text(" - "),
				f.Styled("Don't login as root ^-^", styleWarning),
			)
			ruleStyle = styleWarning
		}
		rule, err := blocks.Rule(blocks.RuleStyle(ruleStyle)).Render(ctx.WithWidth(title.Width()))
		if err != nil {
			return dra.Model{}, err
		}
		return blocks.Model(dra.Stack(dra.AlignLeft, title, rule)).Render(ctx)
	})
}

// palette returns the rows of color swatches.
func palette() beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		f := ctx.Factory()
		cells := make([]dra.Model, len(paletteColors))
		for i, color := range paletteColors {
			cells[i] = f.Styled(swatch, dra.Foreground(color))
		}
		row := dra.Concat(cells...)
		rows := make([]dra.Model, paletteRows)
		for i := range rows {
			rows[i] = row
		}
		return blocks.Model(dra.Stack(dra.AlignLeft, rows...)).Render(ctx)
	})
}

// entry is one row of the info list.
type entry struct {
	key   string
	value dra.Model
}

// keyValue lays the entries of info out one per line as "key: value",
// with the key in the key style. Keys are not padded to a common width,
// which is why blocks.KeyValue, which always aligns the values, is not
// used here.
func keyValue(info sysinfo.Info) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		f := ctx.Factory()
		var lines []dra.Model
		for _, e := range entries(info, f) {
			lines = append(lines, dra.Concat(f.Styled(oneLine(e.key), styleKey), f.Text(": "), e.value))
		}
		return blocks.Model(dra.Stack(dra.AlignLeft, lines...)).Render(ctx)
	})
}

// section produces the entries of one topic of the info list, with
// values built by f. A section may return no entries when it has nothing
// to show.
type section func(sysinfo.Info, *dra.Factory) []entry

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
func entries(info sysinfo.Info, f *dra.Factory) []entry {
	var result []entry
	for _, s := range sections {
		result = append(result, s(info, f)...)
	}
	return result
}

// one returns a single-entry list, the result of most sections.
func one(key string, value dra.Model) []entry {
	return []entry{{key: key, value: value}}
}

// textOrMissing returns s as plain text, or the "Unknown" marker when s
// is blank.
func textOrMissing(f *dra.Factory, s string) dra.Model {
	if strings.TrimSpace(s) == "" {
		return missing(f, "Unknown")
	}
	return f.Text(oneLine(s))
}

// oneLine replaces line breaks in collected text with U+FFFD. beautify
// sanitizes every other control character, but it keeps line breaks as
// layout, and a value that spans several lines would break the list.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return '\uFFFD'
		}
		return r
	}, s)
}

// missing returns reason in the style that marks absent information.
func missing(f *dra.Factory, reason string) dra.Model {
	return f.Styled(reason, styleMissing)
}

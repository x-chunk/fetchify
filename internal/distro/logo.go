package distro

import "strings"

// Run is a piece of a logo line drawn in one color.
type Run struct {
	// Text is the text of the run. It contains no line breaks.
	Text string
	// Color is the index of the color in the palette of the
	// distribution, from 0.
	Color int
}

// Logo is a parsed logo: lines of runs that reference palette colors.
// Lines are not padded; the logo is a left-aligned block. The zero value
// is an empty logo.
type Logo struct {
	lines [][]Run
}

// Size selects one of the two variants of a logo.
type Size uint8

// The logo sizes. Full logos are at most 40 columns by 20 lines, small
// ones at most 16 columns by 8 lines.
const (
	Full Size = iota
	Small
)

// ParseLogo parses a logo in the fetchify format.
//
// The markers ${1} to ${9} switch to the palette color with that number
// for the text that follows, until the next marker; text before the first
// marker uses color 1. A marker of a color beyond colors, the size of the
// palette, falls back to color 1. "$$" is a literal dollar sign, and so is
// a "$" that starts no marker. A trailing newline ends the last line and
// does not start another.
func ParseLogo(data string, colors int) Logo {
	data = strings.TrimSuffix(data, "\n")
	if data == "" {
		return Logo{}
	}
	var logo Logo
	color := 0
	for line := range strings.SplitSeq(data, "\n") {
		var runs []Run
		var text strings.Builder
		flush := func() {
			if text.Len() > 0 {
				runs = append(runs, Run{Text: text.String(), Color: color})
				text.Reset()
			}
		}
		for i := 0; i < len(line); i++ {
			if line[i] != '$' {
				text.WriteByte(line[i])
				continue
			}
			switch {
			case i+1 < len(line) && line[i+1] == '$':
				text.WriteByte('$')
				i++
			case i+3 < len(line) && line[i+1] == '{' && line[i+2] >= '1' && line[i+2] <= '9' && line[i+3] == '}':
				flush()
				color = int(line[i+2] - '1')
				if color >= colors {
					color = 0
				}
				i += 3
			default:
				text.WriteByte('$')
			}
		}
		flush()
		logo.lines = append(logo.lines, runs)
	}
	return logo
}

// Lines returns the runs of every line. A line without text has no runs.
// The caller must not modify the result.
func (l Logo) Lines() [][]Run {
	return l.lines
}

// Width returns the width of the widest line in columns. Logos are
// drawn with characters one column wide each.
func (l Logo) Width() int {
	width := 0
	for _, line := range l.lines {
		n := 0
		for _, r := range line {
			n += len([]rune(r.Text))
		}
		width = max(width, n)
	}
	return width
}

// Height returns the number of lines.
func (l Logo) Height() int {
	return len(l.lines)
}

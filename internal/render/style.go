package render

// Color is one of the eight standard terminal colors. The zero value is
// unset and leaves the terminal's default color in place.
type Color uint8

// The standard terminal colors. Their order matches the SGR color codes:
// the foreground code of a color is 29 plus its value.
const (
	Black Color = iota + 1
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Style is a set of visual properties of a Span. The zero value is the
// terminal's default look. Style is comparable with ==.
type Style struct {
	// FG is the foreground color. The zero value keeps the default.
	FG Color
}

// Foreground returns a Style that only sets the foreground color to c.
func Foreground(c Color) Style {
	return Style{FG: c}
}

// Profile is the level of styling an Engine may emit.
type Profile uint8

const (
	// ProfilePlain emits no escape sequences at all. It is the zero value,
	// so an uninitialized Engine is safe for files and pipes.
	ProfilePlain Profile = iota
	// ProfileANSI16 emits the standard 16-color SGR sequences.
	ProfileANSI16
)

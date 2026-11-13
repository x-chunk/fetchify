package render

import "io"

// Engine encodes blocks as bytes for one Profile. The zero value renders
// plain text. An Engine holds no state between calls and is safe for
// concurrent use.
type Engine struct {
	profile Profile
}

// NewEngine returns an Engine that emits styles at the level of p.
func NewEngine(p Profile) Engine {
	return Engine{profile: p}
}

// Profile returns the profile the Engine renders for.
func (e Engine) Profile() Profile {
	return e.profile
}

// Append appends the encoded form of b to dst and returns the extended
// slice. Lines are separated by '\n'; no trailing newline is added.
//
// Every style switch is a full reset followed by the new style, and the
// style is reset at the end of each line, so a style never leaks across a
// newline or past the end of the output. Unstyled text and empty spans
// produce no escape sequences.
func (e Engine) Append(dst []byte, b Block) []byte {
	for i, line := range b {
		if i > 0 {
			dst = append(dst, '\n')
		}
		// current is the style the terminal is in on this line.
		var current Style
		for _, span := range line {
			if span.text == "" {
				continue
			}
			if style := e.downgrade(span.style); style != current {
				dst = appendSGR(dst, style)
				current = style
			}
			dst = append(dst, span.text...)
		}
		if current != (Style{}) {
			dst = appendSGR(dst, Style{})
		}
	}
	return dst
}

// Render writes the encoded form of b to w with a single Write call and
// returns the number of bytes written and any write error. It makes no
// call for an empty block.
func (e Engine) Render(w io.Writer, b Block) (int, error) {
	buf := e.Append(nil, b)
	if len(buf) == 0 {
		return 0, nil
	}
	return w.Write(buf)
}

// String returns the encoded form of b as a string.
func (e Engine) String(b Block) string {
	return string(e.Append(nil, b))
}

// downgrade reduces s to what the Engine's profile can express.
func (e Engine) downgrade(s Style) Style {
	if e.profile != ProfileANSI16 {
		return Style{}
	}
	return s
}

// appendSGR appends the sequence that switches the terminal to s from any
// previous state: a reset, followed by the foreground color if one is set.
func appendSGR(dst []byte, s Style) []byte {
	dst = append(dst, "\x1b[0"...)
	if s.FG >= Black && s.FG <= White {
		dst = append(dst, ';', '3', '0'+byte(s.FG-Black))
	}
	return append(dst, 'm')
}

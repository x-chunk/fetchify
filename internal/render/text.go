package render

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Span is a run of single-line text with one Style. The zero value is an
// empty span. Create spans with Text and Styled: they sanitize the text,
// which is why the fields are not exported.
type Span struct {
	text  string
	style Style
}

// Text returns an unstyled span holding the sanitized form of s.
func Text(s string) Span {
	return Styled(s, Style{})
}

// Styled returns a span holding the sanitized form of s, drawn with st.
//
// Sanitization replaces every control character (including ESC, newline
// and tab), every bidirectional control and the Unicode line and paragraph
// separators with U+FFFD. A span therefore never spans lines and never
// carries an escape sequence of its own.
func Styled(s string, st Style) Span {
	return Span{text: sanitize(s), style: st}
}

// Width returns the number of terminal cells the span occupies, counting
// one cell per rune.
func (s Span) Width() int {
	return utf8.RuneCountInString(s.text)
}

// Line is a sequence of spans drawn left to right on one terminal line.
// A nil Line is an empty line.
type Line []Span

// Width returns the number of terminal cells the line occupies.
func (l Line) Width() int {
	width := 0
	for _, span := range l {
		width += span.Width()
	}
	return width
}

// String returns the text of the line without any styling.
func (l Line) String() string {
	var sb strings.Builder
	for _, span := range l {
		sb.WriteString(span.text)
	}
	return sb.String()
}

// Block is a sequence of lines drawn top to bottom. A nil Block is empty
// and renders nothing; a Block holding one nil Line is one empty line.
type Block []Line

// Width returns the width of the widest line of the block.
func (b Block) Width() int {
	width := 0
	for _, line := range b {
		width = max(width, line.Width())
	}
	return width
}

// Height returns the number of lines in the block.
func (b Block) Height() int {
	return len(b)
}

// String returns the text of the block without any styling, with a
// newline between lines. It is meant for tests and debugging.
func (b Block) String() string {
	lines := make([]string, len(b))
	for i, line := range b {
		lines[i] = line.String()
	}
	return strings.Join(lines, "\n")
}

// sanitize returns s with every rune that could alter terminal state or
// break the one-line invariant of a Span replaced by U+FFFD. Invalid UTF-8
// bytes decode to U+FFFD as well.
func sanitize(s string) string {
	clean := true
	for _, r := range s {
		if r == utf8.RuneError || unsafeRune(r) {
			clean = false
			break
		}
	}
	if clean {
		return s
	}
	return strings.Map(func(r rune) rune {
		if unsafeRune(r) {
			return utf8.RuneError
		}
		return r
	}, s)
}

// unsafeRune reports whether r must not reach the terminal: C0 and C1
// controls, DEL, bidirectional controls and the line and paragraph
// separators.
func unsafeRune(r rune) bool {
	switch {
	case unicode.IsControl(r):
		return true
	case r >= 0x202A && r <= 0x202E, r >= 0x2066 && r <= 0x2069:
		return true
	case r == 0x2028, r == 0x2029:
		return true
	}
	return false
}

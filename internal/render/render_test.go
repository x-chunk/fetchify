package render

import (
	"bytes"
	"errors"
	"testing"
)

func TestStyledSanitizes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", "hello"},
		{"empty", "", ""},
		{"unicode", "привет ─", "привет ─"},
		{"escape", "a\x1b[31mb", "a�[31mb"},
		{"newline and tab", "a\nb\tc", "a�b�c"},
		{"carriage return", "a\rb", "a�b"},
		{"c1 control", "a\u009bb", "a�b"},
		{"bidi override", "a‮b", "a�b"},
		{"line separator", "a b", "a�b"},
		{"invalid utf-8", "a\xffb", "a�b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Line{Text(tt.in)}).String(); got != tt.want {
				t.Errorf("Text(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestWidthAndHeight(t *testing.T) {
	block := Block{
		{Text("ab"), Styled("─█", Foreground(Red))},
		nil,
		{Text("abcde")},
	}
	if got := block[0].Width(); got != 4 {
		t.Errorf("line width = %d, want 4", got)
	}
	if got := block.Width(); got != 5 {
		t.Errorf("block width = %d, want 5", got)
	}
	if got := block.Height(); got != 3 {
		t.Errorf("block height = %d, want 3", got)
	}
	if got := (Block)(nil).Width(); got != 0 {
		t.Errorf("empty block width = %d, want 0", got)
	}
}

func TestStack(t *testing.T) {
	a := Block{{Text("a")}}
	b := Block{{Text("b")}, {Text("c")}}

	if got := Stack(a, nil, b).String(); got != "a\nb\nc" {
		t.Errorf("Stack = %q", got)
	}
	if got := Stack(); got.Height() != 0 {
		t.Errorf("Stack() height = %d, want 0", got.Height())
	}
	if a.Height() != 1 || b.Height() != 2 {
		t.Error("Stack modified its arguments")
	}
}

func TestJoin(t *testing.T) {
	left := Block{{Text("ab")}, {Text("a")}}
	right := Block{{Text("x")}, {Text("y")}, {Text("z")}}

	tests := []struct {
		name   string
		gap    int
		blocks []Block
		want   string
	}{
		{"none", 1, nil, ""},
		{"single", 1, []Block{left}, "ab\na"},
		{"shorter left is padded", 1, []Block{left, right}, "ab x\na  y\n   z"},
		{"shorter right is not padded", 1, []Block{right, left}, "x ab\ny a\nz "},
		{"negative gap", -3, []Block{left, right}, "abx\na y\n  z"},
		{"empty blocks are skipped", 2, []Block{nil, left, nil, right}, "ab  x\na   y\n    z"},
		{"only empty blocks", 1, []Block{nil, {}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Join(tt.gap, tt.blocks...).String(); got != tt.want {
				t.Errorf("Join = %q, want %q", got, tt.want)
			}
		})
	}
	if left.String() != "ab\na" {
		t.Error("Join modified its arguments")
	}
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		unit  string
		width int
		want  string
	}{
		{"─", 3, "───"},
		{"-=", 5, "-=-=-"},
		{"─", 0, ""},
		{"─", -1, ""},
		{"", 4, ""},
	}
	for _, tt := range tests {
		got := Repeat(tt.unit, tt.width, Style{})
		if got.text != tt.want {
			t.Errorf("Repeat(%q, %d) = %q, want %q", tt.unit, tt.width, got.text, tt.want)
		}
	}
}

func TestEngineAppend(t *testing.T) {
	red := Foreground(Red)
	green := Foreground(Green)

	tests := []struct {
		name    string
		profile Profile
		block   Block
		want    string
	}{
		{"empty block", ProfileANSI16, nil, ""},
		{"unstyled", ProfileANSI16, Block{{Text("hi")}}, "hi"},
		{"styled", ProfileANSI16, Block{{Styled("hi", red)}}, "\x1b[0;31mhi\x1b[0m"},
		{"styled at plain", ProfilePlain, Block{{Styled("hi", red)}}, "hi"},
		{
			"equal neighbors share one switch", ProfileANSI16,
			Block{{Styled("a", red), Styled("b", red)}},
			"\x1b[0;31mab\x1b[0m",
		},
		{
			"switch between styles", ProfileANSI16,
			Block{{Styled("a", red), Styled("b", green), Text("c")}},
			"\x1b[0;31ma\x1b[0;32mb\x1b[0mc",
		},
		{
			"style is reset before a newline", ProfileANSI16,
			Block{{Styled("a", red)}, {Styled("b", red)}},
			"\x1b[0;31ma\x1b[0m\n\x1b[0;31mb\x1b[0m",
		},
		{
			"empty styled span emits nothing", ProfileANSI16,
			Block{{Styled("", red)}, {Text("a")}},
			"\na",
		},
		{"empty lines", ProfileANSI16, Block{nil, nil}, "\n"},
		{
			"out-of-range color is ignored", ProfileANSI16,
			Block{{Styled("a", Foreground(Color(200)))}},
			"\x1b[0ma\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEngine(tt.profile).String(tt.block); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngineZeroValueIsPlain(t *testing.T) {
	var e Engine
	if got := e.String(Block{{Styled("hi", Foreground(Red))}}); got != "hi" {
		t.Errorf("got %q, want %q", got, "hi")
	}
	if e.Profile() != ProfilePlain {
		t.Errorf("Profile() = %d, want ProfilePlain", e.Profile())
	}
}

// countingWriter records how many times Write is called.
type countingWriter struct {
	bytes.Buffer
	calls int
	err   error
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.err != nil {
		return 0, w.err
	}
	return w.Buffer.Write(p)
}

func TestEngineRender(t *testing.T) {
	e := NewEngine(ProfilePlain)

	w := &countingWriter{}
	n, err := e.Render(w, Block{{Text("a")}, {Text("b")}})
	if err != nil || n != 3 || w.calls != 1 || w.String() != "a\nb" {
		t.Errorf("Render = %d, %v with %d calls and %q", n, err, w.calls, w.String())
	}

	w = &countingWriter{}
	if n, err := e.Render(w, nil); err != nil || n != 0 || w.calls != 0 {
		t.Errorf("Render(empty) = %d, %v with %d calls", n, err, w.calls)
	}

	failure := errors.New("write failed")
	w = &countingWriter{err: failure}
	if _, err := e.Render(w, Block{{Text("a")}}); !errors.Is(err, failure) {
		t.Errorf("Render error = %v, want %v", err, failure)
	}
}

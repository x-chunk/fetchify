package distro

import (
	"reflect"
	"testing"
)

func TestParseLogo(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		colors int
		want   [][]Run
	}{
		{"default color", "ab\ncd\n", 2, [][]Run{{{"ab", 0}}, {{"cd", 0}}}},
		{"markers", "${2}a${1}b\nc", 2, [][]Run{{{"a", 1}, {"b", 0}}, {{"c", 0}}}},
		{"color carries over lines", "${2}a\nb", 2, [][]Run{{{"a", 1}}, {{"b", 1}}}},
		{"missing color falls back", "${3}a", 2, [][]Run{{{"a", 0}}}},
		{"dollars", "$$5 $x ${0} ${a} $", 1, [][]Run{{{"$5 $x ${0} ${a} $", 0}}}},
		{"empty line", "a\n\nb", 1, [][]Run{{{"a", 0}}, nil, {{"b", 0}}}},
		{"marker only", "${2}", 2, [][]Run{nil}},
		{"empty", "", 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLogo(tt.data, tt.colors).Lines(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogoSize(t *testing.T) {
	l := ParseLogo("${1}ab${2}cd\nx\n", 2)
	if l.Width() != 4 || l.Height() != 2 {
		t.Errorf("size = %dx%d, want 4x2", l.Width(), l.Height())
	}
	if (Logo{}).Width() != 0 || (Logo{}).Height() != 0 {
		t.Error("zero logo is not empty")
	}
}

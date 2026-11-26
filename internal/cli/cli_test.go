package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		args []string
		want Options
	}{
		{nil, Options{}},
		{[]string{"--help"}, Options{Help: true}},
		{[]string{"-h"}, Options{Help: true}},
		{[]string{"--dont-clear"}, Options{DontClear: true}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.args)
		if err != nil || got != tt.want {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", tt.args, got, err, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, args := range [][]string{{"--nope"}, {"extra"}, {"--help=maybe"}} {
		_, err := Parse(args)
		if !errors.Is(err, ErrUsage) {
			t.Errorf("Parse(%q) error = %v, want ErrUsage", args, err)
		}
	}
}

func TestUsage(t *testing.T) {
	if !strings.HasPrefix(Usage(), "Usage: fetchify") {
		t.Errorf("usage = %q", Usage())
	}
}

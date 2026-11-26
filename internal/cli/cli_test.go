package cli

import (
	"errors"
	"strings"
	"testing"

	"go.xchunk.org/fetchify/internal/config"
)

// logos are the logo names of the tests.
var logos = []string{"arch", "koala"}

func TestParse(t *testing.T) {
	tests := []struct {
		args []string
		want Options
	}{
		{nil, Options{}},
		{[]string{"--help"}, Options{Help: true}},
		{[]string{"-h"}, Options{Help: true}},
		{[]string{"--dont-clear"}, Options{DontClear: true}},
		{[]string{"--config", "/x.toml", "--logo=arch", "--no-icons"}, Options{ConfigPath: "/x.toml", Logo: "arch", NoIcons: true}},
		{[]string{"--logo", "none", "--print-config"}, Options{Logo: "none", PrintConfig: true}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.args, logos)
		if err != nil || got != tt.want {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", tt.args, got, err, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, args := range [][]string{{"--nope"}, {"extra"}, {"--help=maybe"}, {"--logo", "windows"}, {"--config"}} {
		_, err := Parse(args, logos)
		if !errors.Is(err, ErrUsage) {
			t.Errorf("Parse(%q) error = %v, want ErrUsage", args, err)
		}
	}
}

func TestUnknownLogoListsKnown(t *testing.T) {
	_, err := Parse([]string{"--logo", "windows"}, logos)
	if err == nil || !strings.Contains(err.Error(), "arch, koala") {
		t.Errorf("err = %v", err)
	}
}

func TestApply(t *testing.T) {
	cfg := config.Defaults()
	Options{Logo: "koala", NoIcons: true}.Apply(&cfg)
	if cfg.Logo.Name != "koala" || cfg.Icons.Enabled {
		t.Errorf("Apply: %+v %+v", cfg.Logo, cfg.Icons)
	}
	cfg = config.Defaults()
	Options{}.Apply(&cfg)
	if cfg.Logo.Name != config.LogoAuto || !cfg.Icons.Enabled {
		t.Error("empty options changed the configuration")
	}
}

func TestUsage(t *testing.T) {
	if !strings.HasPrefix(Usage(), "Usage: fetchify") {
		t.Errorf("usage = %q", Usage())
	}
}

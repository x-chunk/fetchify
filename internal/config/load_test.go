package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// known is the set of identifiers the tests validate against.
var known = Known{
	Items: []string{"os", "kernel", "uptime", "packages", "shell", "terminal", "de", "wm",
		"host", "cpu", "gpu", "display", "battery", "memory", "swap", "disks", "public_ip", "local_ip"},
	Logos: []string{"arch", "ubuntu", "koala", "linux"},
}

// parse parses a configuration given as text.
func parse(t *testing.T, text string) (Config, error) {
	t.Helper()
	return Parse("config.toml", []byte(text), known)
}

// problems returns the problems of err as strings, or fails.
func problems(t *testing.T, err error) []string {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("error %v is not an *Error", err)
	}
	var out []string
	for _, p := range e.Problems {
		out = append(out, p.String())
	}
	return out
}

func TestMergeRules(t *testing.T) {
	cfg, err := parse(t, "")
	if err != nil || !reflect.DeepEqual(cfg, Defaults()) {
		t.Errorf("empty file: %v\n%+v", err, cfg)
	}

	cfg, err = parse(t, "[bars]\nwidth = 20\n")
	want := Defaults()
	want.Bars.Width = 20
	if err != nil || !reflect.DeepEqual(cfg, want) {
		t.Errorf("one key: %v\n%+v", err, cfg)
	}

	cfg, err = parse(t, "[[group]]\ntitle = \"Mine\"\nitems = [\"cpu\"]\n")
	if err != nil || len(cfg.Groups) != 1 || cfg.Groups[0].Title != "Mine" || cfg.Groups[0].Icon != IconAuto {
		t.Errorf("one group replaces the defaults: %v\n%+v", err, cfg.Groups)
	}

	cfg, err = parse(t, "group = []\n")
	if err != nil || cfg.Groups == nil || len(cfg.Groups) != 0 {
		t.Errorf("no groups: %v %#v", err, cfg.Groups)
	}

	cfg, err = parse(t, "[logo]\ncolors = [\"#fff\"]\n[items.cpu]\nlabel = \"Processor\"\nicon = \"\"\ntimeout_ms = 50\n")
	if err != nil {
		t.Fatal(err)
	}
	cpu := cfg.Items["cpu"]
	if *cpu.Label != "Processor" || *cpu.Icon != "" || cpu.Timeout != 50*time.Millisecond {
		t.Errorf("item settings: %+v", cpu)
	}
	if !reflect.DeepEqual(cfg.Logo.Colors, []string{"#fff"}) || !reflect.DeepEqual(cfg.Items["disks"], Defaults().Items["disks"]) {
		t.Errorf("arrays replace, other items keep defaults: %+v", cfg)
	}
}

func TestTypeAndKeyProblems(t *testing.T) {
	_, err := parse(t, strings.Join([]string{
		`clear = "yes"`,
		`colour = 1`,
		`[border]`,
		`style = 3`,
		`shadow = true`,
		`[bars]`,
		`thresholds = [1]`,
		`[[group]]`,
		`title = "A"`,
		`items = ["os", 2]`,
		`[items.disks]`,
		`ignore = "/boot"`,
	}, "\n"))
	want := []string{
		"1:1: clear: expected a boolean, got a string",
		"2:1: unknown key colour",
		"4:1: border.style: expected a string, got an integer",
		"5:1: unknown key border.shadow",
		"7:1: bars.thresholds: expected an array of two integers",
		"10:1: group.0.items: expected an array of strings, got an integer",
		"12:1: items.disks.ignore: expected an array of strings, got a string",
	}
	if got := problems(t, err); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestSyntaxError(t *testing.T) {
	_, err := parse(t, "a = 1\nb = [\n")
	got := problems(t, err)
	if len(got) != 1 || !strings.HasPrefix(got[0], "2:") {
		t.Errorf("got %q", got)
	}
	if !strings.HasPrefix(err.Error(), "config.toml:2:") {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "none.toml"), known)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
	path := filepath.Join(t.TempDir(), "c.toml")
	if err := os.WriteFile(path, []byte("footer = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cfg, err := Load(path, known); err != nil || cfg.Footer {
		t.Errorf("Load = %+v, %v", cfg.Footer, err)
	}
}

func TestDefaultPath(t *testing.T) {
	env := func(v string) func(string) string { return func(string) string { return v } }
	tests := []struct {
		xdg, home, want string
	}{
		{"/xdg", "/home/a", "/xdg/fetchify/config.toml"},
		{"", "/home/a", "/home/a/.config/fetchify/config.toml"},
		{"relative/dir", "/home/a", "/home/a/.config/fetchify/config.toml"},
		{"", "", ""},
	}
	for _, tt := range tests {
		if got := DefaultPath(env(tt.xdg), tt.home); got != tt.want {
			t.Errorf("DefaultPath(%q, %q) = %q, want %q", tt.xdg, tt.home, got, tt.want)
		}
	}
}

package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/v2/internal/cli"
	"go.xchunk.org/fetchify/v2/internal/config"
	"go.xchunk.org/fetchify/v2/internal/distro"
)

// failingWriter rejects every write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "--print-config") {
		t.Errorf("help does not mention --dont-clear: %q", stdout.String())
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--nope"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want it empty", stdout.String())
	}
}

func TestRunPrintsScreen(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--dont-clear"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr %q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "collected in") || !strings.HasSuffix(out, "\n") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRunWriteError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var stderr bytes.Buffer
	if code := run([]string{"--dont-clear"}, failingWriter{}, &stderr); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "disk full") {
		t.Errorf("stderr = %q, want the write error", stderr.String())
	}
}

// brokenConfig writes an invalid configuration file and points the
// default location at it.
func brokenConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "fetchify", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("colour = 1\n[bars]\nwidth = 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunInvalidConfig(t *testing.T) {
	path := brokenConfig(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--dont-clear"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	want := "fetchify: " + path + ":1:1: unknown key colour\n" +
		"fetchify: " + path + ":3:1: bars.width: 99 is outside 4-40\n"
	if stderr.String() != want || stdout.Len() != 0 {
		t.Errorf("stderr = %q, stdout = %q", stderr.String(), stdout.String())
	}
}

func TestHelpAndPrintConfigIgnoreBrokenConfig(t *testing.T) {
	brokenConfig(t)
	for _, flag := range []string{"--help", "--print-config", "--version"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{flag}, &stdout, &stderr); code != 0 || stdout.Len() == 0 {
			t.Errorf("%s: exit %d, stderr %q", flag, code, stderr.String())
		}
	}
	var stdout bytes.Buffer
	run([]string{"--print-config"}, &stdout, io.Discard)
	if !bytes.Equal(stdout.Bytes(), config.DefaultTOML()) {
		t.Error("--print-config does not print the default file")
	}
}

func TestLoadConfig(t *testing.T) {
	empty := func(string) string { return "" }
	missing := filepath.Join(t.TempDir(), "none.toml")
	if _, err := loadConfig(cli.Options{ConfigPath: missing}, empty); err == nil {
		t.Error("a missing --config file is not an error")
	}
	dir := t.TempDir()
	xdg := func(k string) string {
		if k == "XDG_CONFIG_HOME" {
			return dir
		}
		return ""
	}
	cfg, err := loadConfig(cli.Options{}, xdg)
	if err != nil || !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Errorf("missing default file: %v", err)
	}
	printed := filepath.Join(dir, "printed.toml")
	if err := os.WriteFile(printed, config.DefaultTOML(), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = loadConfig(cli.Options{ConfigPath: printed}, empty)
	if err != nil || !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Errorf("the printed configuration is not the default one: %v", err)
	}
}

func TestRunFlagOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.toml")
	if err := os.WriteFile(path, []byte("[logo]\nname = \"arch\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts, _ := cli.Parse([]string{"--config", path, "--logo", "none"}, distro.IDs())
	cfg, err := loadConfig(opts, os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	opts.Apply(&cfg)
	if cfg.Logo.Name != "none" {
		t.Errorf("logo = %q, want the flag to win", cfg.Logo.Name)
	}
}

func TestRunNotATerminal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	var stdout bytes.Buffer
	if code := run([]string{"--clear"}, &stdout, io.Discard); code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if bytes.ContainsRune(stdout.Bytes(), '\x1b') {
		t.Errorf("output to a buffer contains escape sequences: %q", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	if code := run([]string{"--version"}, &stdout, io.Discard); code != 0 || stdout.String() != "fetchify dev\n" {
		t.Errorf("exit %d, output %q", code, stdout.String())
	}
}

func TestIsTerminal(t *testing.T) {
	if isTerminal(&bytes.Buffer{}) {
		t.Error("a buffer is a terminal")
	}
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if isTerminal(f) {
		t.Error("a regular file is a terminal")
	}
}

func TestComposerOptions(t *testing.T) {
	if composerOptions(config.ModeAuto) != nil {
		t.Error("auto fixes a profile")
	}
	for mode := range profiles {
		if len(composerOptions(mode)) != 1 {
			t.Errorf("%s does not fix the profile", mode)
		}
	}
	var buf bytes.Buffer
	c := beautify.NewComposer(&buf, composerOptions(config.ModeTrueColor)...)
	if c.Profile() != dra.ProfileTrueColor {
		t.Errorf("profile = %v", c.Profile())
	}
}

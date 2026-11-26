package sysinfo

import (
	"context"
	"errors"
	"slices"
	"testing"
)

// env returns a getenv function backed by vars.
func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// commands returns a runner that answers with the canned output for the
// command name and fails for every other command.
func commands(outputs map[string]string) runner {
	return func(name string, _ ...string) ([]byte, error) {
		out, ok := outputs[name]
		if !ok {
			return nil, errors.New("command not found")
		}
		return []byte(out), nil
	}
}

func TestParseMemoryTypes(t *testing.T) {
	out := "Physical Memory Array\n" +
		"\tUse: System Memory\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"\tType Detail: Synchronous\n" +
		"Memory Device\n" +
		"\tType: Unknown\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"Memory Device\n" +
		"\tType: Other\n" +
		"Memory Device\n" +
		"\tType: RAM\n" +
		"Memory Device\n" +
		"\tType: DDR5\n"
	want := []string{"DDR4", "DDR5"}
	if got := parseMemoryTypes(out); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIsIgnoredMount(t *testing.T) {
	tests := []struct {
		mount string
		want  bool
	}{
		{"/", false},
		{"/home", false},
		{"/boot", true},
		{"/boot/efi", true},
		{"/var/lib/docker", true},
		{"/variable", false},
		{"/opt/run", false},
	}
	for _, tt := range tests {
		if got := isIgnoredMount(tt.mount); got != tt.want {
			t.Errorf("isIgnoredMount(%q) = %v, want %v", tt.mount, got, tt.want)
		}
	}
}

// TestCollect is the smoke test against the machine that runs the tests.
func TestCollect(t *testing.T) {
	info := Collect(context.Background(), SystemSource(), Facts(), Options{})
	if info.CPU.State != Present || info.CPU.V.Count < 1 {
		t.Errorf("CPU = %+v, want at least one logical CPU", info.CPU)
	}
}

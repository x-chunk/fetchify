package sysinfo

import (
	"context"
	"errors"
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

// TestCollect is the smoke test against the machine that runs the tests.
func TestCollect(t *testing.T) {
	info := Collect(context.Background(), SystemSource(), Facts(), Options{})
	if info.CPU.State != Present || info.CPU.V.Count < 1 {
		t.Errorf("CPU = %+v, want at least one logical CPU", info.CPU)
	}
}

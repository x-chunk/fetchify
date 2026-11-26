package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testdata", "case.golden")
	got := []byte("\x1b[31mred\x1b[0m\n")

	err := check(path, got, false)
	if err == nil || !strings.Contains(err.Error(), "-update") {
		t.Fatalf("missing file: err = %v, want a hint to run with -update", err)
	}

	if err := check(path, got, true); err != nil {
		t.Fatalf("update: %v", err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != string(got) {
		t.Fatalf("written file = %q, %v; want %q", data, err, got)
	}

	if err := check(path, got, false); err != nil {
		t.Errorf("equal bytes: %v", err)
	}

	err = check(path, []byte("\x1b[32mred\x1b[0m\n"), false)
	if err == nil {
		t.Fatal("different bytes: no error")
	}
	if !strings.Contains(err.Error(), `\x1b[32m`) || !strings.Contains(err.Error(), `\x1b[31m`) {
		t.Errorf("mismatch message does not quote both values: %v", err)
	}
}

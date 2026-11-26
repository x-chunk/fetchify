package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// failingWriter rejects every write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "--dont-clear") {
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
	var stderr bytes.Buffer
	if code := run([]string{"--dont-clear"}, failingWriter{}, &stderr); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "disk full") {
		t.Errorf("stderr = %q, want the write error", stderr.String())
	}
}

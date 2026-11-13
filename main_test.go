package main

import (
	"bytes"
	"strings"
	"testing"
)

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
	if !strings.Contains(out, "==== Fetchify ====") || !strings.HasSuffix(out, "\n") {
		t.Errorf("unexpected output: %q", out)
	}
}

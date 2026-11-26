package sysinfo

import (
	"context"
	"io/fs"
	"testing"
	"time"
)

func TestSystemSource(t *testing.T) {
	src := SystemSource()
	if _, err := fs.Stat(src.FS, "proc/self"); err != nil {
		t.Errorf("FS is not rooted at /: %v", err)
	}
	st, err := src.Statfs("/")
	if err != nil || st.Total == 0 || st.Available > st.Total {
		t.Errorf("Statfs(/) = %+v, %v", st, err)
	}
	if src.PID <= 0 {
		t.Errorf("PID = %d", src.PID)
	}
	if name, err := src.LookupUser(src.UID); err != nil || name == "" {
		t.Errorf("LookupUser(%d) = %q, %v", src.UID, name, err)
	}
}

func TestRunCommandHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := runCommand(ctx, "sleep", "5"); err == nil {
		t.Error("a killed command reported success")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("runCommand returned after %v, want it to stop at the deadline", elapsed)
	}
}

func TestRunCommandMissingTool(t *testing.T) {
	if _, err := runCommand(context.Background(), "fetchify-no-such-tool"); err == nil {
		t.Error("a missing tool reported success")
	}
}

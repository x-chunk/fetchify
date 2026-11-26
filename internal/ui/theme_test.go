package ui

import (
	"testing"

	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
)

func TestNewTheme(t *testing.T) {
	arch, _ := distro.ByID("arch")
	cfg := config.Defaults()

	th := newTheme(cfg, arch, false)
	if th.accent != arch.Accent || th.frame != dra.Foreground(arch.Accent) || th.title.Attrs&dra.Bold == 0 {
		t.Errorf("auto accent: %+v", th)
	}
	if th.caption.FG.IsSet() {
		t.Errorf("caption of a user is colored: %+v", th.caption)
	}
	if th.value != (dra.Style{}) {
		t.Errorf("value style = %+v, want the terminal's own foreground", th.value)
	}

	cfg.Colors.Accent = "#ff0000"
	if th := newTheme(cfg, arch, true); th.accent != dra.RGB(255, 0, 0) || th.caption.FG != dra.Red {
		t.Errorf("custom accent and root: %+v", th)
	}
}

package ui

import (
	"testing"

	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/v2/internal/config"
	"go.xchunk.org/fetchify/v2/internal/distro"
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

func TestFrameColors(t *testing.T) {
	ubuntu, _ := distro.ByID("ubuntu")
	cfg := config.Defaults()
	if th := newTheme(cfg, ubuntu, false); !th.gradient.IsZero() {
		t.Error("accent frame has a gradient")
	}
	cfg.Border.Color = config.BorderLogo
	// Ubuntu has one real color; TerminalDefault cannot be a stop.
	if th := newTheme(cfg, ubuntu, false); !th.gradient.IsZero() || th.frame != dra.Foreground(ubuntu.Colors[0]) {
		t.Errorf("logo frame of a one-color logo: %+v", th)
	}
	cfg.Border.Color = config.BorderRainbow
	if th := newTheme(cfg, ubuntu, false); th.gradient.IsZero() {
		t.Error("rainbow frame has no gradient")
	}
}

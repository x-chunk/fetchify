package ui

import (
	"testing"

	"go.xchunk.org/fetchify/internal/golden"
	"go.xchunk.org/fetchify/internal/render"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// screen renders the whole screen for info the way main writes it: at
// profile p and followed by a newline.
func screen(info sysinfo.Info, p render.Profile) []byte {
	return append(render.NewEngine(p).Append(nil, Build(info)), '\n')
}

// TestScreenGolden pins the exact bytes of the screen, escape sequences
// and padding included.
func TestScreenGolden(t *testing.T) {
	root := sample()
	root.User = "root"
	root.Root = true

	fixtures := []struct {
		name string
		info sysinfo.Info
	}{
		{"sample", sample()},
		{"empty", sysinfo.Info{}},
		{"root", root},
	}
	profiles := []struct {
		name    string
		profile render.Profile
	}{
		{"plain", render.ProfilePlain},
		{"ansi16", render.ProfileANSI16},
	}
	for _, f := range fixtures {
		for _, p := range profiles {
			t.Run(f.name+"_"+p.name, func(t *testing.T) {
				golden.Assert(t, f.name+"_"+p.name, screen(f.info, p.profile))
			})
		}
	}
}

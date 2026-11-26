package ui

import (
	"testing"

	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/golden"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

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
		profile dra.Profile
	}{
		{"plain", dra.ProfilePlain},
		{"ansi16", dra.ProfileANSI16},
	}
	for _, f := range fixtures {
		for _, p := range profiles {
			t.Run(f.name+"_"+p.name, func(t *testing.T) {
				golden.Assert(t, f.name+"_"+p.name, append(render(t, Build(f.info), p.profile), '\n'))
			})
		}
	}
}

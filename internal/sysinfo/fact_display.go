package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// FactDisplay is the resolution of every connected output.
const FactDisplay Fact = "display"

// Display is a connected output.
type Display struct {
	// Width and Height are the preferred resolution in pixels.
	Width, Height int
}

func init() {
	collectors[FactDisplay] = fact(collectDisplay, func(i *Info) *Value[[]Display] { return &i.Display })
}

// collectDisplay reads the connectors of the DRM cards. The first mode a
// connector lists is its preferred one. The fact is Absent when no output
// is connected, Unknown when sysfs has no DRM class at all.
func collectDisplay(_ context.Context, s *session) Value[[]Display] {
	entries, err := fs.ReadDir(s.fsys(), "sys/class/drm")
	if err != nil {
		return Value[[]Display]{}
	}
	var displays []Display
	for _, e := range entries {
		card, connector, found := strings.Cut(e.Name(), "-")
		if !found || !cardName.MatchString(card) || connector == "" {
			continue
		}
		dir := path.Join("sys/class/drm", e.Name())
		if status, _ := readTrimmed(s.fsys(), path.Join(dir, "status")); status != "connected" {
			continue
		}
		modes, _ := readTrimmed(s.fsys(), path.Join(dir, "modes"))
		first, _, _ := strings.Cut(modes, "\n")
		if d, ok := parseMode(first); ok {
			displays = append(displays, d)
		}
	}
	if len(displays) == 0 {
		return absent[[]Display]()
	}
	return present(displays)
}

// parseMode parses a DRM mode name such as "2560x1440". Interlaced modes
// carry an "i" suffix, which is ignored.
func parseMode(mode string) (Display, bool) {
	w, h, found := strings.Cut(strings.TrimSuffix(strings.TrimSpace(mode), "i"), "x")
	width, err1 := strconv.Atoi(w)
	height, err2 := strconv.Atoi(h)
	if !found || err1 != nil || err2 != nil || width <= 0 || height <= 0 {
		return Display{}, false
	}
	return Display{Width: width, Height: height}, true
}

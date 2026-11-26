package ui

import (
	"testing"
	"unicode"

	"go.xchunk.org/fetchify/internal/config"
)

func TestIconsArePrivateUse(t *testing.T) {
	check := func(name string, r rune) {
		if !unicode.Is(unicode.Co, r) {
			t.Errorf("%s: %U is not in a private-use range", name, r)
		}
	}
	for id, r := range itemIcons {
		check(id, r)
		if _, ok := registry[id]; !ok {
			t.Errorf("icon for unknown item %q", id)
		}
	}
	for title, r := range groupIcons {
		check(title, r)
	}
	for _, g := range config.Defaults().Groups {
		if _, ok := groupIcons[g.Title]; !ok {
			t.Errorf("default group %q has no icon", g.Title)
		}
		for _, id := range g.Items {
			if _, ok := itemIcons[id]; !ok && id != "os" {
				t.Errorf("default item %q has no icon", id)
			}
		}
	}
}

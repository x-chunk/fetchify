package ui

import (
	"slices"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// Facts returns the facts that the screen of cfg shows, sorted. The user
// and the host name are always needed for the caption and os-release for
// the logo and the accent; other facts only when an item of a group
// needs them. Pass the result to sysinfo.Collect.
func Facts(cfg config.Config) []sysinfo.Fact {
	set := map[sysinfo.Fact]bool{
		sysinfo.FactUser:     true,
		sysinfo.FactHostname: true,
		sysinfo.FactOS:       true,
	}
	for _, g := range cfg.Groups {
		for _, id := range g.Items {
			for _, f := range registry[id].facts {
				set[f] = true
			}
			if id == "memory" && cfg.Items["memory"].ShowType {
				set[sysinfo.FactMemoryType] = true
			}
		}
	}
	facts := make([]sysinfo.Fact, 0, len(set))
	for f := range set {
		facts = append(facts, f)
	}
	slices.Sort(facts)
	return facts
}

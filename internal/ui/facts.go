package ui

import (
	"slices"
	"time"

	"go.xchunk.org/fetchify/v2/internal/config"
	"go.xchunk.org/fetchify/v2/internal/sysinfo"
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

// Options returns the collection settings of cfg: the default time limit,
// the limits of single items applied to their facts, and the settings of
// the disks and public_ip items. When two items set a limit for the same
// fact, the longer one wins.
func Options(cfg config.Config) sysinfo.Options {
	opts := sysinfo.Options{
		Timeout:     cfg.Timeout,
		Timeouts:    map[sysinfo.Fact]time.Duration{},
		DiskIgnore:  cfg.Items["disks"].Ignore,
		PublicIPURL: cfg.Items["public_ip"].URL,
	}
	for id, settings := range cfg.Items {
		if settings.Timeout <= 0 {
			continue
		}
		for _, f := range registry[id].facts {
			opts.Timeouts[f] = max(opts.Timeouts[f], settings.Timeout)
		}
	}
	return opts
}

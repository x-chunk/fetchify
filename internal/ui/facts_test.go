package ui

import (
	"slices"
	"testing"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

func TestFacts(t *testing.T) {
	known := sysinfo.Facts()
	cfg := config.Defaults()
	facts := Facts(cfg)
	for _, f := range facts {
		if !slices.Contains(known, f) {
			t.Errorf("unknown fact %q", f)
		}
	}
	for _, want := range []sysinfo.Fact{sysinfo.FactUser, sysinfo.FactOS, sysinfo.FactDisks, sysinfo.FactSession} {
		if !slices.Contains(facts, want) {
			t.Errorf("default facts lack %q", want)
		}
	}
	for _, optIn := range []sysinfo.Fact{sysinfo.FactPublicIP, sysinfo.FactMemoryType, sysinfo.FactCPUUsage} {
		if slices.Contains(facts, optIn) {
			t.Errorf("default facts collect opt-in %q", optIn)
		}
	}
	cfg.Groups = []config.Group{{Title: "X", Items: []string{"memory", "public_ip"}}}
	cfg.Items["memory"] = config.Item{ShowType: true}
	want := []sysinfo.Fact{sysinfo.FactHostname, sysinfo.FactMemory, sysinfo.FactMemoryType, sysinfo.FactOS, sysinfo.FactPublicIP, sysinfo.FactUser}
	if got := Facts(cfg); !slices.Equal(got, want) {
		t.Errorf("Facts = %v, want %v", got, want)
	}
}

func TestEveryItemFactIsKnown(t *testing.T) {
	known := sysinfo.Facts()
	for id, it := range registry {
		if len(it.facts) == 0 {
			t.Errorf("item %q needs no facts", id)
		}
		for _, f := range it.facts {
			if !slices.Contains(known, f) {
				t.Errorf("item %q needs unknown fact %q", id, f)
			}
		}
	}
}

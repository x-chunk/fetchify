package ui

import (
	"slices"
	"testing"
	"time"

	"go.xchunk.org/fetchify/v2/internal/config"
	"go.xchunk.org/fetchify/v2/internal/sysinfo"
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

func TestOptions(t *testing.T) {
	cfg := config.Defaults()
	cfg.Timeout = 300 * time.Millisecond
	cfg.Items["public_ip"] = config.Item{URL: "https://example.test", Timeout: 2 * time.Second}
	cfg.Items["wm"] = config.Item{Timeout: 50 * time.Millisecond}
	cfg.Items["de"] = config.Item{Timeout: 80 * time.Millisecond}
	opts := Options(cfg)
	if opts.Timeout != 300*time.Millisecond || opts.PublicIPURL != "https://example.test" {
		t.Errorf("options = %+v", opts)
	}
	if !slices.Equal(opts.DiskIgnore, cfg.Items["disks"].Ignore) {
		t.Errorf("disk ignore = %v", opts.DiskIgnore)
	}
	if opts.Timeouts[sysinfo.FactPublicIP] != 2*time.Second {
		t.Errorf("public_ip timeout = %v", opts.Timeouts[sysinfo.FactPublicIP])
	}
	// wm and de both need the wm fact; the longer limit wins.
	if opts.Timeouts[sysinfo.FactWM] != 80*time.Millisecond || opts.Timeouts[sysinfo.FactSession] != 50*time.Millisecond {
		t.Errorf("shared fact timeouts = %v", opts.Timeouts)
	}
	cfg.Items["disks"] = config.Item{Ignore: []string{}}
	if got := Options(cfg).DiskIgnore; got == nil || len(got) != 0 {
		t.Errorf("an empty ignore list must stay empty, not nil: %#v", got)
	}
}

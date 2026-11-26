package sysinfo

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestCollectSystem is the smoke test against the machine that runs the
// tests. It checks only facts every Linux machine has.
func TestCollectSystem(t *testing.T) {
	info := Collect(context.Background(), SystemSource(), Facts(), Options{Timeout: 2 * time.Second})
	checks := map[Fact]State{
		FactUser:     info.User.State,
		FactHostname: info.Hostname.State,
		FactKernel:   info.Kernel.State,
		FactUptime:   info.Uptime.State,
		FactCPU:      info.CPU.State,
		FactMemory:   info.Memory.State,
	}
	for f, state := range checks {
		if state != Present {
			t.Errorf("%s is %v, want present", f, state)
		}
	}
	if info.CPU.V.Count < 1 {
		t.Errorf("CPU = %+v, want at least one logical CPU", info.CPU.V)
	}
}

// TestCollectFixturesConcurrently collects every fact from every fixture
// machine many times in parallel, for the race detector.
func TestCollectFixturesConcurrently(t *testing.T) {
	machines := []machine{ubuntuDesktop, archLaptop, sshServer, container, raspberryPi, btrfsBox, placeholderDMI}
	var wg sync.WaitGroup
	for range 10 {
		for _, m := range machines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				Collect(context.Background(), m.source(), Facts(), Options{})
			}()
		}
	}
	wg.Wait()
}

func TestFactsSorted(t *testing.T) {
	facts := Facts()
	if len(facts) == 0 {
		t.Fatal("no facts registered")
	}
	for i := 1; i < len(facts); i++ {
		if facts[i-1] >= facts[i] {
			t.Errorf("Facts() not sorted at %d: %q, %q", i, facts[i-1], facts[i])
		}
	}
}

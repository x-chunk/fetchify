package sysinfo

import (
	"context"
	"testing"
	"testing/fstest"
)

const (
	kib = 1 << 10
	gib = 1 << 30
)

func TestMemory(t *testing.T) {
	meminfo := func(s string) Source {
		return Source{FS: fstest.MapFS{"proc/meminfo": {Data: []byte(s)}}}
	}
	tests := []struct {
		name string
		src  Source
		want Value[Usage]
	}{
		{"fixture", ubuntuDesktop.source(), present(Usage{Used: 8 * gib, Total: 32 * gib, Available: 24 * gib})},
		{"old kernel", meminfo("MemTotal: 1000 kB\nMemFree: 100 kB\nBuffers: 50 kB\nCached: 250 kB\n"),
			present(Usage{Used: 600 * kib, Total: 1000 * kib, Available: 400 * kib})},
		{"available above total", meminfo("MemTotal: 10 kB\nMemAvailable: 20 kB\n"),
			present(Usage{Used: 0, Total: 10 * kib, Available: 10 * kib})},
		{"no total", meminfo("MemFree: 1 kB\n"), Value[Usage]{}},
		{"missing", Source{}, Value[Usage]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactMemory}, Options{}).Memory
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseMeminfo(t *testing.T) {
	got := parseMeminfo("MemTotal:  2 kB\nHugePages_Total:  3\nbad\nX: y kB\n")
	if got["MemTotal"] != 2*kib || got["HugePages_Total"] != 3 || len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

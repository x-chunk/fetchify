package sysinfo

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
)

// FactMemory is the use of physical memory.
const FactMemory Fact = "memory"

// Usage is an amount of a resource in use out of a total, in bytes.
type Usage struct {
	// Used is the amount in use.
	Used uint64
	// Total is the size of the resource.
	Total uint64
	// Available is what an unprivileged user can still use. For file
	// systems it is less than Total-Used because of blocks reserved for
	// root, and percentages follow df by computing Used/(Used+Available).
	Available uint64
}

func init() {
	collectors[FactMemory] = fact(collectMemory, func(i *Info) *Value[Usage] { return &i.Memory })
}

// collectMemory reads proc/meminfo. Memory in use is what is not
// available: page cache and reclaimable buffers do not count.
func collectMemory(_ context.Context, s *session) Value[Usage] {
	m, ok := readMeminfo(s.fsys())
	total := m["MemTotal"]
	if !ok || total == 0 {
		return Value[Usage]{}
	}
	available, ok := m["MemAvailable"]
	if !ok {
		// Kernels before 3.14 have no MemAvailable.
		available = m["MemFree"] + m["Buffers"] + m["Cached"]
	}
	available = min(available, total)
	return present(Usage{Used: total - available, Total: total, Available: available})
}

// readMeminfo reads proc/meminfo.
func readMeminfo(fsys fs.FS) (map[string]uint64, bool) {
	data, err := fs.ReadFile(fsys, "proc/meminfo")
	if err != nil {
		return nil, false
	}
	return parseMeminfo(string(data)), true
}

// parseMeminfo parses lines such as "MemTotal:  32768000 kB" into sizes
// in bytes. Lines that do not parse are skipped.
func parseMeminfo(data string) map[string]uint64 {
	values := map[string]uint64{}
	for line := range strings.SplitSeq(data, "\n") {
		key, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		if len(fields) > 1 && fields[1] == "kB" {
			n *= 1024
		}
		values[key] = n
	}
	return values
}

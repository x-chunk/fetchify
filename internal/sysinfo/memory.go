package sysinfo

import (
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v3/mem"
)

// memory returns the state of physical memory, or nil when it cannot be
// read. Module types come from dmidecode, which only works for root, so
// they are looked up only when root is true.
func memory(run runner, root bool) *Memory {
	stat, err := mem.VirtualMemory()
	if err != nil || stat.Total == 0 {
		return nil
	}
	result := &Memory{
		Usage: Usage{Used: stat.Used, Total: stat.Total, Percent: stat.UsedPercent},
	}
	if root {
		if out, err := run("dmidecode", "-t", "memory"); err == nil {
			result.Types = parseMemoryTypes(string(out))
		}
	}
	return result
}

// parseMemoryTypes extracts the distinct module types from the output of
// `dmidecode -t memory`, in order of first appearance. Empty slots report
// "Unknown" or "Other" and are skipped, as are the "RAM" entries of the
// array-level records.
func parseMemoryTypes(out string) []string {
	var types []string
	for line := range strings.SplitSeq(out, "\n") {
		value, found := strings.CutPrefix(strings.TrimSpace(line), "Type:")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" ||
			strings.Contains(value, "Unknown") ||
			strings.Contains(value, "Other") ||
			strings.Contains(value, "RAM") {
			continue
		}
		if !slices.Contains(types, value) {
			types = append(types, value)
		}
	}
	return types
}

// swap returns the swap usage, or nil when it cannot be read.
func swap() *Usage {
	stat, err := mem.SwapMemory()
	if err != nil {
		return nil
	}
	return &Usage{Used: stat.Used, Total: stat.Total, Percent: stat.UsedPercent}
}

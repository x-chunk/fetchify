package sysinfo

import (
	"context"
	"slices"
	"strings"
)

// FactMemoryType is the list of installed memory module types.
const FactMemoryType Fact = "memory_type"

func init() {
	collectors[FactMemoryType] = fact(collectMemoryType, func(i *Info) *Value[[]string] { return &i.MemoryType })
}

// collectMemoryType asks dmidecode for the memory modules. dmidecode
// reads the firmware tables through /dev/mem or sysfs files that only
// root can read, so the fact is Absent for other users. It is Unknown
// when dmidecode is missing or fails.
func collectMemoryType(ctx context.Context, s *session) Value[[]string] {
	if s.src.UID != 0 {
		return absent[[]string]()
	}
	out, err := s.run(ctx, "dmidecode", "-t", "memory")
	if err != nil {
		return Value[[]string]{}
	}
	types := parseMemoryTypes(string(out))
	if len(types) == 0 {
		return Value[[]string]{}
	}
	return present(types)
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

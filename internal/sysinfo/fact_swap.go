package sysinfo

import "context"

// FactSwap is the use of swap space.
const FactSwap Fact = "swap"

func init() {
	collectors[FactSwap] = fact(collectSwap, func(i *Info) *Value[Usage] { return &i.Swap })
}

// collectSwap reads proc/meminfo. The fact is Absent when no swap is
// configured.
func collectSwap(_ context.Context, s *session) Value[Usage] {
	m, ok := readMeminfo(s.fsys())
	total, hasTotal := m["SwapTotal"]
	if !ok || !hasTotal {
		return Value[Usage]{}
	}
	if total == 0 {
		return absent[Usage]()
	}
	free := min(m["SwapFree"], total)
	return present(Usage{Used: total - free, Total: total, Available: free})
}

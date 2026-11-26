package ui

import "go.xchunk.org/fetchify/internal/sysinfo"

// facts lists what the screen shows, so that nothing else is collected.
var facts = []sysinfo.Fact{
	sysinfo.FactUser,
	sysinfo.FactHostname,
	sysinfo.FactKernel,
	sysinfo.FactOS,
	sysinfo.FactUptime,
	sysinfo.FactShell,
	sysinfo.FactSession,
	sysinfo.FactWM,
	sysinfo.FactCPU,
	sysinfo.FactGPU,
	sysinfo.FactMemory,
	sysinfo.FactMemoryType,
	sysinfo.FactSwap,
	sysinfo.FactDisks,
}

// Facts returns the facts the screen needs. Pass them to sysinfo.Collect.
func Facts() []sysinfo.Fact {
	return append([]sysinfo.Fact(nil), facts...)
}

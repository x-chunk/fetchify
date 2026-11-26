package ui

import "go.xchunk.org/fetchify/internal/sysinfo"

// facts lists what the screen shows, so that nothing else is collected.
var facts = []sysinfo.Fact{
	sysinfo.FactUser,
	sysinfo.FactHostname,
	sysinfo.FactKernel,
}

// Facts returns the facts the screen needs. Pass them to sysinfo.Collect.
func Facts() []sysinfo.Fact {
	return append([]sysinfo.Fact(nil), facts...)
}

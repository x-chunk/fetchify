package sysinfo

import "context"

// FactHostname is the host name reported by the kernel.
const FactHostname Fact = "hostname"

func init() {
	collectors[FactHostname] = fact(collectHostname, func(i *Info) *Value[string] { return &i.Hostname })
}

// collectHostname reads the host name of the UTS namespace, which is
// also what a container reports. An empty name is Unknown.
func collectHostname(_ context.Context, s *session) Value[string] {
	name, err := readTrimmed(s.fsys(), "proc/sys/kernel/hostname")
	if err != nil || name == "" {
		return Value[string]{}
	}
	return present(name)
}

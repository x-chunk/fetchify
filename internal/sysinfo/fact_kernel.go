package sysinfo

import "context"

// FactKernel is the release of the running kernel.
const FactKernel Fact = "kernel"

func init() {
	collectors[FactKernel] = fact(collectKernel, func(i *Info) *Value[string] { return &i.Kernel })
}

// collectKernel reads the kernel release, for example "6.8.0-45-generic".
func collectKernel(_ context.Context, s *session) Value[string] {
	release, err := readTrimmed(s.fsys(), "proc/sys/kernel/osrelease")
	if err != nil || release == "" {
		return Value[string]{}
	}
	return present(release)
}

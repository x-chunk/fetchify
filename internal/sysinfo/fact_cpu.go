package sysinfo

import (
	"context"
	"io/fs"
	"strings"
)

// FactCPU is the processor model and the number of logical CPUs.
const FactCPU Fact = "cpu"

// CPU describes the processor.
type CPU struct {
	// Model is the model name as the kernel reports it, for example
	// "AMD Ryzen 7 5800X 8-Core Processor". It is not cleaned for
	// display and can be empty.
	Model string
	// Count is the number of logical CPUs.
	Count int
}

func init() {
	collectors[FactCPU] = fact(collectCPU, func(i *Info) *Value[CPU] { return &i.CPU })
}

// collectCPU reads proc/cpuinfo.
func collectCPU(_ context.Context, s *session) Value[CPU] {
	data, err := fs.ReadFile(s.fsys(), "proc/cpuinfo")
	if err != nil {
		return Value[CPU]{}
	}
	cpu := parseCPUInfo(string(data))
	if cpu.Count == 0 && cpu.Model == "" {
		return Value[CPU]{}
	}
	return present(cpu)
}

// parseCPUInfo extracts the model and the logical CPU count from
// proc/cpuinfo. The model is the first "model name" line. ARM kernels
// often have none and describe the chip in "Hardware" or the board in
// "Model" instead, which are used in that order.
func parseCPUInfo(data string) CPU {
	var cpu CPU
	var hardware, model string
	for line := range strings.SplitSeq(data, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "processor":
			cpu.Count++
		case "model name":
			if cpu.Model == "" {
				cpu.Model = value
			}
		case "Hardware":
			if hardware == "" {
				hardware = value
			}
		case "Model":
			if model == "" {
				model = value
			}
		}
	}
	for _, fallback := range []string{hardware, model} {
		if cpu.Model == "" {
			cpu.Model = fallback
		}
	}
	return cpu
}

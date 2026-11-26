package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strings"
)

// Facts about the GPUs beyond their names.
const (
	// FactGPUDriver is the kernel driver of every GPU.
	FactGPUDriver Fact = "gpu_driver"
	// FactGPUMemory is the use of video memory of every GPU that reports
	// it.
	FactGPUMemory Fact = "gpu_memory"
)

// mib is the unit nvidia-smi reports memory in.
const mib = 1 << 20

func init() {
	collectors[FactGPUDriver] = fact(collectGPUDriver, func(i *Info) *Value[[]string] { return &i.GPUDriver })
	collectors[FactGPUMemory] = fact(collectGPUMemory, func(i *Info) *Value[[]Usage] { return &i.GPUMemory })
}

// collectGPUDriver reads the driver of every GPU from its uevent file, in
// bus order. GPUs without a bound driver are left out. The fact is Absent
// without GPUs and Unknown when no driver is known.
func collectGPUDriver(_ context.Context, s *session) Value[[]string] {
	devices := drmDevices(s.fsys())
	if len(devices) == 0 {
		return absent[[]string]()
	}
	var drivers []string
	for _, d := range devices {
		data, err := fs.ReadFile(s.fsys(), path.Join(d.dir, "uevent"))
		if err != nil {
			continue
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			if driver, found := strings.CutPrefix(line, "DRIVER="); found && driver != "" {
				drivers = append(drivers, strings.TrimSpace(driver))
			}
		}
	}
	if len(drivers) == 0 {
		return Value[[]string]{}
	}
	return present(drivers)
}

// collectGPUMemory reads the video memory of every GPU in bus order: from
// the amdgpu counters in sysfs, and through nvidia-smi for NVIDIA GPUs.
// GPUs without counters, such as integrated ones, are left out. The fact
// is Absent without GPUs and Unknown when no GPU reports its memory.
func collectGPUMemory(ctx context.Context, s *session) Value[[]Usage] {
	devices := drmDevices(s.fsys())
	if len(devices) == 0 {
		return absent[[]Usage]()
	}
	var nvidia [][]float64
	var asked bool
	var usages []Usage
	for _, d := range devices {
		total, ok1 := readInt(s.fsys(), path.Join(d.dir, "mem_info_vram_total"))
		used, ok2 := readInt(s.fsys(), path.Join(d.dir, "mem_info_vram_used"))
		if ok1 && ok2 && total > 0 && used >= 0 {
			usages = append(usages, vramUsage(uint64(used), uint64(total)))
			continue
		}
		if d.vendor != "10de" {
			continue
		}
		if !asked {
			asked = true
			nvidia = nvidiaSMI(ctx, s, "memory.used,memory.total")
		}
		if len(nvidia) > 0 {
			row := nvidia[0]
			nvidia = nvidia[1:]
			if row[0] >= 0 && row[1] > 0 {
				usages = append(usages, vramUsage(uint64(row[0])*mib, uint64(row[1])*mib))
			}
		}
	}
	if len(usages) == 0 {
		return Value[[]Usage]{}
	}
	return present(usages)
}

// vramUsage builds a Usage from used and total video memory.
func vramUsage(used, total uint64) Usage {
	used = min(used, total)
	return Usage{Used: used, Total: total, Available: total - used}
}

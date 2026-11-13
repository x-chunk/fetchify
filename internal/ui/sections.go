package ui

import (
	"fmt"
	"strings"

	"go.xchunk.org/fetchify/internal/render"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

func osSection(info sysinfo.Info) []entry {
	name := info.OS.Name
	if name == "arch" {
		name = "i use arch, btw"
	}
	return one("OS", textOrMissing(strings.TrimSpace(name+" "+info.OS.Version)))
}

func kernelSection(info sysinfo.Info) []entry {
	return one("Kernel", textOrMissing(info.Kernel))
}

func uptimeSection(info sysinfo.Info) []entry {
	return one("Uptime", render.Text(info.Uptime.String()))
}

func shellSection(info sysinfo.Info) []entry {
	return one("Shell", textOrMissing(info.Shell))
}

func desktopSection(info sysinfo.Info) []entry {
	if info.Desktop.WM == "" {
		return one("WM", missing("Unknown"))
	}
	return one("WM", render.Text(fmt.Sprintf("%s (%s)", info.Desktop.WM, info.Desktop.Server)))
}

func cpuSection(info sysinfo.Info) []entry {
	return one("CPU",
		textOrMissing(info.CPU.Model),
		render.Text(fmt.Sprintf(" (%d cores)", info.CPU.Cores)),
	)
}

// gpuSection returns one entry per display adapter, or a single "Unknown"
// entry when none is known.
func gpuSection(info sysinfo.Info) []entry {
	if len(info.GPUs) == 0 {
		return one("GPU", missing("Unknown"))
	}
	var result []entry
	for _, name := range info.GPUs {
		result = append(result, one("GPU", render.Text(name))...)
	}
	return result
}

func memorySection(info sysinfo.Info) []entry {
	if info.Memory == nil {
		return one("RAM", missing("Error"))
	}
	return one("RAM", usage(info.Memory.Usage, strings.Join(info.Memory.Types, ", "))...)
}

func swapSection(info sysinfo.Info) []entry {
	switch {
	case info.Swap == nil:
		return one("Swap", missing("Error"))
	case info.Swap.Total == 0:
		return one("Swap", missing("Disabled"))
	}
	return one("Swap", usage(*info.Swap, "")...)
}

// diskSection returns one entry per disk and no entries when there are
// no disks to show.
func diskSection(info sysinfo.Info) []entry {
	var result []entry
	for _, disk := range info.Disks {
		key := fmt.Sprintf("Disk (%s)", disk.Mountpoint)
		result = append(result, one(key, usage(disk.Usage, disk.Filesystem)...)...)
	}
	return result
}

// bytesPerGiB converts a byte count to the unit shown on screen.
const bytesPerGiB = 1 << 30

// usage formats u as "<used> GiB / <total> GiB (<percent>%)", followed by
// " - <category>" when category is not empty.
func usage(u sysinfo.Usage, category string) []render.Span {
	spans := []render.Span{
		render.Styled(fmt.Sprintf("%.2f GiB", float64(u.Used)/bytesPerGiB), styleUsed),
		render.Text(" / "),
		render.Styled(fmt.Sprintf("%.2f GiB", float64(u.Total)/bytesPerGiB), styleTotal),
		render.Text(fmt.Sprintf(" (%.0f%%)", u.Percent)),
	}
	if category != "" {
		spans = append(spans, render.Text(" - "), render.Styled(category, styleCategory))
	}
	return spans
}

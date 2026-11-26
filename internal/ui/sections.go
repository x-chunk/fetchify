package ui

import (
	"fmt"
	"strings"

	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/distro"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

func osSection(info sysinfo.Info, f *dra.Factory) []entry {
	release := distro.ParseRelease(info.OS.V)
	name := release.ID
	if name == "arch" {
		name = "i use arch, btw"
	}
	return one("OS", textOrMissing(f, strings.TrimSpace(name+" "+release.VersionID)))
}

func kernelSection(info sysinfo.Info, f *dra.Factory) []entry {
	return one("Kernel", textOrMissing(f, info.Kernel.V))
}

func uptimeSection(info sysinfo.Info, f *dra.Factory) []entry {
	return one("Uptime", f.Text(info.Uptime.V.String()))
}

func shellSection(info sysinfo.Info, f *dra.Factory) []entry {
	return one("Shell", textOrMissing(f, info.Shell.V))
}

func desktopSection(info sysinfo.Info, f *dra.Factory) []entry {
	if info.Desktop.WM == "" {
		return one("WM", missing(f, "Unknown"))
	}
	return one("WM", f.Text(oneLine(fmt.Sprintf("%s (%s)", info.Desktop.WM, info.Desktop.Server))))
}

func cpuSection(info sysinfo.Info, f *dra.Factory) []entry {
	return one("CPU", dra.Concat(
		textOrMissing(f, info.CPU.Model),
		f.Text(fmt.Sprintf(" (%d cores)", info.CPU.Cores)),
	))
}

// gpuSection returns one entry per display adapter, or a single "Unknown"
// entry when none is known.
func gpuSection(info sysinfo.Info, f *dra.Factory) []entry {
	if len(info.GPUs) == 0 {
		return one("GPU", missing(f, "Unknown"))
	}
	var result []entry
	for _, name := range info.GPUs {
		result = append(result, one("GPU", f.Text(oneLine(name)))...)
	}
	return result
}

func memorySection(info sysinfo.Info, f *dra.Factory) []entry {
	if info.Memory == nil {
		return one("RAM", missing(f, "Error"))
	}
	return one("RAM", usage(f, info.Memory.Usage, strings.Join(info.Memory.Types, ", ")))
}

func swapSection(info sysinfo.Info, f *dra.Factory) []entry {
	switch {
	case info.Swap == nil:
		return one("Swap", missing(f, "Error"))
	case info.Swap.Total == 0:
		return one("Swap", missing(f, "Disabled"))
	}
	return one("Swap", usage(f, *info.Swap, ""))
}

// diskSection returns one entry per disk and no entries when there are
// no disks to show.
func diskSection(info sysinfo.Info, f *dra.Factory) []entry {
	var result []entry
	for _, disk := range info.Disks {
		key := fmt.Sprintf("Disk (%s)", disk.Mountpoint)
		result = append(result, one(key, usage(f, disk.Usage, disk.Filesystem))...)
	}
	return result
}

// bytesPerGiB converts a byte count to the unit shown on screen.
const bytesPerGiB = 1 << 30

// usage formats u as "<used> GiB / <total> GiB (<percent>%)", followed by
// " - <category>" when category is not empty.
func usage(f *dra.Factory, u sysinfo.Usage, category string) dra.Model {
	m := dra.Concat(
		f.Styled(fmt.Sprintf("%.2f GiB", float64(u.Used)/bytesPerGiB), styleUsed),
		f.Text(" / "),
		f.Styled(fmt.Sprintf("%.2f GiB", float64(u.Total)/bytesPerGiB), styleTotal),
		f.Text(fmt.Sprintf(" (%.0f%%)", u.Percent)),
	)
	if category != "" {
		m = dra.Concat(m, f.Text(" - "), f.Styled(oneLine(category), styleCategory))
	}
	return m
}

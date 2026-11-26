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
	if info.WM.State != sysinfo.Present {
		return one("WM", missing(f, "Unknown"))
	}
	if info.Session.State != sysinfo.Present {
		return one("WM", f.Text(oneLine(info.WM.V)))
	}
	return one("WM", f.Text(oneLine(fmt.Sprintf("%s (%s)", info.WM.V, info.Session.V))))
}

func cpuSection(info sysinfo.Info, f *dra.Factory) []entry {
	return one("CPU", dra.Concat(
		textOrMissing(f, info.CPU.V.Model),
		f.Text(fmt.Sprintf(" (%d cores)", info.CPU.V.Count)),
	))
}

// gpuSection returns one entry per display adapter, or a single "Unknown"
// entry when none is known.
func gpuSection(info sysinfo.Info, f *dra.Factory) []entry {
	if len(info.GPU.V) == 0 {
		return one("GPU", missing(f, "Unknown"))
	}
	var result []entry
	for _, name := range info.GPU.V {
		result = append(result, one("GPU", f.Text(oneLine(name)))...)
	}
	return result
}

func memorySection(info sysinfo.Info, f *dra.Factory) []entry {
	if info.Memory.State != sysinfo.Present {
		return one("RAM", missing(f, "Error"))
	}
	types := strings.Join(info.MemoryType.V, ", ")
	return one("RAM", usage(f, info.Memory.V, share(info.Memory.V.Used, info.Memory.V.Total), types))
}

func swapSection(info sysinfo.Info, f *dra.Factory) []entry {
	switch info.Swap.State {
	case sysinfo.Unknown:
		return one("Swap", missing(f, "Error"))
	case sysinfo.Absent:
		return one("Swap", missing(f, "Disabled"))
	}
	return one("Swap", usage(f, info.Swap.V, share(info.Swap.V.Used, info.Swap.V.Total), ""))
}

// diskSection returns one entry per disk and no entries when there are
// no disks to show.
func diskSection(info sysinfo.Info, f *dra.Factory) []entry {
	var result []entry
	for _, disk := range info.Disks.V {
		key := fmt.Sprintf("Disk (%s)", disk.Mountpoint)
		percent := share(disk.Usage.Used, disk.Usage.Used+disk.Usage.Available)
		result = append(result, one(key, usage(f, disk.Usage, percent, disk.Filesystem))...)
	}
	return result
}

// bytesPerGiB converts a byte count to the unit shown on screen.
const bytesPerGiB = 1 << 30

// share returns part as a percentage of whole, or 0 when whole is 0.
func share(part, whole uint64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}

// usage formats u as "<used> GiB / <total> GiB (<percent>%)", followed by
// " - <category>" when category is not empty.
func usage(f *dra.Factory, u sysinfo.Usage, percent float64, category string) dra.Model {
	m := dra.Concat(
		f.Styled(fmt.Sprintf("%.2f GiB", float64(u.Used)/bytesPerGiB), styleUsed),
		f.Text(" / "),
		f.Styled(fmt.Sprintf("%.2f GiB", float64(u.Total)/bytesPerGiB), styleTotal),
		f.Text(fmt.Sprintf(" (%.0f%%)", percent)),
	)
	if category != "" {
		m = dra.Concat(m, f.Text(" - "), f.Styled(oneLine(category), styleCategory))
	}
	return m
}

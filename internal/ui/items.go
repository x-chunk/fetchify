package ui

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"go.xchunk.org/fetchify/v2/internal/config"
	"go.xchunk.org/fetchify/v2/internal/distro"
	"go.xchunk.org/fetchify/v2/internal/sysinfo"
)

// itemEnv is what an item needs besides the collected facts.
type itemEnv struct {
	cfg    config.Config
	distro distro.Distro
}

// item is one configurable entry of a group.
type item struct {
	id string
	// label is the default label.
	label string
	// facts are what the item needs collected.
	facts []sysinfo.Fact
	// rows builds the rows of the item. It returns no rows when the facts
	// are Absent, which hides the item, and one row with a muted
	// "unknown" when they are Unknown.
	rows func(info sysinfo.Info, env itemEnv) []row
}

// registry holds every item by identifier.
var registry = map[string]item{}

// register adds items to the registry.
func register(items ...item) {
	for _, it := range items {
		registry[it.id] = it
	}
}

// ItemIDs returns the identifiers of all items, sorted. Pass them to
// config as the known items.
func ItemIDs() []string {
	return slices.Sorted(maps.Keys(registry))
}

// text returns the rows of an item whose value is one string.
func text(label string, v sysinfo.Value[string]) []row {
	switch v.State {
	case sysinfo.Present:
		return []row{{label: label, value: textValue(v.V)}}
	case sysinfo.Absent:
		return nil
	}
	return []row{unknownRow(label)}
}

func init() {
	register(
		item{id: "os", label: "OS", facts: []sysinfo.Fact{sysinfo.FactOS}, rows: osRows},
		item{id: "kernel", label: "Kernel", facts: []sysinfo.Fact{sysinfo.FactKernel},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Kernel", info.Kernel) }},
		item{id: "uptime", label: "Uptime", facts: []sysinfo.Fact{sysinfo.FactUptime}, rows: uptimeRows},
		item{id: "packages", label: "Packages", facts: []sysinfo.Fact{sysinfo.FactPackages}, rows: packagesRows},
		item{id: "shell", label: "Shell", facts: []sysinfo.Fact{sysinfo.FactShell},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Shell", info.Shell) }},
		item{id: "terminal", label: "Terminal", facts: []sysinfo.Fact{sysinfo.FactTerminal},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Terminal", info.Terminal) }},
		item{id: "de", label: "DE", facts: []sysinfo.Fact{sysinfo.FactDE, sysinfo.FactWM}, rows: deRows},
		item{id: "wm", label: "WM", facts: []sysinfo.Fact{sysinfo.FactWM, sysinfo.FactSession}, rows: wmRows},
		item{id: "host", label: "Host", facts: []sysinfo.Fact{sysinfo.FactHost},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Host", info.Host) }},
		item{id: "cpu", label: "CPU", facts: []sysinfo.Fact{sysinfo.FactCPU}, rows: cpuRows},
		item{id: "gpu", label: "GPU", facts: []sysinfo.Fact{sysinfo.FactGPU}, rows: gpuRows},
		item{id: "display", label: "Display", facts: []sysinfo.Fact{sysinfo.FactDisplay}, rows: displayRows},
		item{id: "battery", label: "Battery", facts: []sysinfo.Fact{sysinfo.FactBattery}, rows: batteryRows},
		item{id: "memory", label: "RAM", facts: []sysinfo.Fact{sysinfo.FactMemory}, rows: memoryRows},
		item{id: "swap", label: "Swap", facts: []sysinfo.Fact{sysinfo.FactSwap}, rows: swapRows},
		item{id: "disks", label: "Disk", facts: []sysinfo.Fact{sysinfo.FactDisks}, rows: diskRows},
	)
}

// osRows shows the name of the distribution with its icon.
func osRows(info sysinfo.Info, env itemEnv) []row {
	rows := osRowsWithoutIcon(info)
	for i := range rows {
		rows[i].icon = env.distro.Icon
	}
	return rows
}

func osRowsWithoutIcon(info sysinfo.Info) []row {
	if info.OS.State != sysinfo.Present {
		return []row{unknownRow("OS")}
	}
	r := distro.ParseRelease(info.OS.V)
	name := r.PrettyName
	if name == "" {
		name = strings.TrimSpace(r.Name + " " + r.VersionID)
	}
	if name == "" {
		return []row{unknownRow("OS")}
	}
	value := textValue(name)
	if r.ID == "arch" {
		value = append(value, segment{text: " btw", role: roleMuted})
	}
	return []row{{label: "OS", value: value}}
}

func uptimeRows(info sysinfo.Info, _ itemEnv) []row {
	if info.Uptime.State != sysinfo.Present {
		return []row{unknownRow("Uptime")}
	}
	return []row{{label: "Uptime", value: textValue(formatUptime(info.Uptime.V))}}
}

// packagesRows shows "1842 (dpkg), 12 (flatpak)" with the manager names
// as details.
func packagesRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Packages.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("Packages")}
	}
	var value []segment
	for i, p := range info.Packages.V {
		if i > 0 {
			value = append(value, segment{text: ", "})
		}
		value = append(value,
			segment{text: strconv.Itoa(p.Count)},
			segment{text: " (" + p.Manager + ")", role: roleMuted, detail: true})
	}
	return []row{{label: "Packages", value: value}}
}

// deRows hides the desktop environment when it only repeats the window
// manager, as on a bare sway or Hyprland session.
func deRows(info sysinfo.Info, _ itemEnv) []row {
	if info.DE.State == sysinfo.Present && info.WM.State == sysinfo.Present && strings.EqualFold(info.DE.V, info.WM.V) {
		return nil
	}
	return text("DE", info.DE)
}

func wmRows(info sysinfo.Info, _ itemEnv) []row {
	rows := text("WM", info.WM)
	if len(rows) == 1 && info.WM.State == sysinfo.Present && info.Session.State == sysinfo.Present {
		rows[0].badge = &badge{text: string(info.Session.V)}
	}
	return rows
}

func cpuRows(info sysinfo.Info, _ itemEnv) []row {
	if info.CPU.State != sysinfo.Present || info.CPU.V.Model == "" {
		return []row{unknownRow("CPU")}
	}
	value := textValue(cleanCPU(info.CPU.V.Model))
	if n := info.CPU.V.Count; n > 0 {
		unit := "cores"
		if n == 1 {
			unit = "core"
		}
		value = append(value, segment{text: fmt.Sprintf("  %d %s", n, unit), role: roleMuted, detail: true})
	}
	return []row{{label: "CPU", value: value}}
}

func gpuRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.GPU.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("GPU")}
	}
	rows := make([]row, len(info.GPU.V))
	for i, name := range info.GPU.V {
		rows[i] = row{label: "GPU", value: textValue(cleanGPU(name))}
	}
	return rows
}

func displayRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Display.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("Display")}
	}
	modes := make([]string, len(info.Display.V))
	for i, d := range info.Display.V {
		modes[i] = fmt.Sprintf("%dx%d", d.Width, d.Height)
	}
	return []row{{label: "Display", value: textValue(strings.Join(modes, ", "))}}
}

// batteryRows draws one usage row per battery with a status badge. A
// discharging battery below 20 percent gets the warning look.
func batteryRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Battery.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("Battery")}
	}
	rows := make([]row, len(info.Battery.V))
	for i, b := range info.Battery.V {
		percent := clampPercent(float64(b.Percent))
		look := badgeNeutral
		switch {
		case b.Status == "Charging" || b.Status == "Full":
			look = badgeGood
		case b.Status == "Discharging" && percent < 20:
			look = badgeWarning
		}
		rows[i] = row{
			label: "Battery",
			meter: &meter{percent: percent, inverted: true},
			badge: &badge{text: strings.ToLower(b.Status), look: look},
		}
	}
	return rows
}

func memoryRows(info sysinfo.Info, env itemEnv) []row {
	if info.Memory.State != sysinfo.Present {
		return []row{unknownRow("RAM")}
	}
	r := row{label: "RAM", meter: sizeMeter(info.Memory.V, info.Memory.V.Total)}
	if env.cfg.Items["memory"].ShowType && info.MemoryType.State == sysinfo.Present {
		r.value = []segment{{text: strings.Join(info.MemoryType.V, ", "), role: roleMuted, detail: true}}
	}
	return []row{r}
}

// swapRows shows a muted "disabled" badge when there is no swap.
func swapRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Swap.State {
	case sysinfo.Absent:
		return []row{{label: "Swap", badge: &badge{text: "disabled"}}}
	case sysinfo.Unknown:
		return []row{unknownRow("Swap")}
	}
	return []row{{label: "Swap", meter: sizeMeter(info.Swap.V, info.Swap.V.Total)}}
}

// diskRows draws one usage row per disk, labeled with the mount point
// and with the file system as a detail. The percentage counts only the
// space users can use, as df does.
func diskRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Disks.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("Disk")}
	}
	rows := make([]row, len(info.Disks.V))
	for i, d := range info.Disks.V {
		rows[i] = row{
			label: d.Mountpoint,
			meter: sizeMeter(d.Usage, d.Usage.Used+d.Usage.Available),
			value: []segment{{text: d.Filesystem, role: roleMuted, detail: true}},
		}
	}
	return rows
}

// sizeMeter builds a meter with sizes. base is what the percentage is
// relative to.
func sizeMeter(u sysinfo.Usage, base uint64) *meter {
	return &meter{percent: percentOf(u.Used, base), used: u.Used, total: u.Total, sizes: true}
}

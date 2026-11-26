package ui

import (
	"fmt"
	"strconv"
	"strings"

	"go.xchunk.org/fetchify/v2/internal/sysinfo"
)

// Temperatures in degrees Celsius where the color of a reading turns to
// warning and danger.
const (
	tempWarning = 70
	tempDanger  = 85
)

// dateLayout is the format of the date item.
const dateLayout = "2006-01-02 15:04"

func init() {
	register(
		item{id: "local_ip", label: "Local IP", facts: []sysinfo.Fact{sysinfo.FactLocalIP}, rows: localIPRows},
		item{id: "public_ip", label: "Public IP", facts: []sysinfo.Fact{sysinfo.FactPublicIP},
			rows: func(info sysinfo.Info, _ itemEnv) []row {
				return stringer("Public IP", info.PublicIP.State, info.PublicIP.V.String())
			}},
		item{id: "network", label: "Network", facts: []sysinfo.Fact{sysinfo.FactNetwork}, rows: networkRows},
		item{id: "bluetooth", label: "Bluetooth", facts: []sysinfo.Fact{sysinfo.FactBluetooth}, rows: bluetoothRows},
		item{id: "cpu_temp", label: "CPU Temp", facts: []sysinfo.Fact{sysinfo.FactCPUTemp}, rows: cpuTempRows},
		item{id: "gpu_temp", label: "GPU Temp", facts: []sysinfo.Fact{sysinfo.FactGPUTemp}, rows: gpuTempRows},
		item{id: "cpu_usage", label: "CPU Usage", facts: []sysinfo.Fact{sysinfo.FactCPUUsage}, rows: cpuUsageRows},
		item{id: "load", label: "Load", facts: []sysinfo.Fact{sysinfo.FactLoad}, rows: loadRows},
		item{id: "gpu_driver", label: "GPU Driver", facts: []sysinfo.Fact{sysinfo.FactGPUDriver}, rows: gpuDriverRows},
		item{id: "gpu_memory", label: "VRAM", facts: []sysinfo.Fact{sysinfo.FactGPUMemory}, rows: gpuMemoryRows},
		item{id: "board", label: "Board", facts: []sysinfo.Fact{sysinfo.FactBoard},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Board", info.Board) }},
		item{id: "bios", label: "BIOS", facts: []sysinfo.Fact{sysinfo.FactBIOS}, rows: biosRows},
		item{id: "init", label: "Init", facts: []sysinfo.Fact{sysinfo.FactInit},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Init", info.Init) }},
		item{id: "processes", label: "Processes", facts: []sysinfo.Fact{sysinfo.FactProcesses},
			rows: func(info sysinfo.Info, _ itemEnv) []row {
				return stringer("Processes", info.Processes.State, strconv.Itoa(info.Processes.V))
			}},
		item{id: "users", label: "Users", facts: []sysinfo.Fact{sysinfo.FactUsers},
			rows: func(info sysinfo.Info, _ itemEnv) []row {
				return stringer("Users", info.Users.State, strconv.Itoa(info.Users.V))
			}},
		item{id: "locale", label: "Locale", facts: []sysinfo.Fact{sysinfo.FactLocale},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Locale", info.Locale) }},
		item{id: "timezone", label: "Timezone", facts: []sysinfo.Fact{sysinfo.FactTimezone},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Timezone", info.Timezone) }},
		item{id: "date", label: "Date", facts: []sysinfo.Fact{sysinfo.FactDate},
			rows: func(info sysinfo.Info, _ itemEnv) []row {
				return stringer("Date", info.Date.State, info.Date.V.Format(dateLayout))
			}},
		item{id: "theme", label: "Theme", facts: []sysinfo.Fact{sysinfo.FactTheme},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Theme", info.Theme) }},
		item{id: "icon_theme", label: "Icons", facts: []sysinfo.Fact{sysinfo.FactIconTheme},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Icons", info.IconTheme) }},
		item{id: "font", label: "Font", facts: []sysinfo.Fact{sysinfo.FactFont},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Font", info.Font) }},
		item{id: "cursor", label: "Cursor", facts: []sysinfo.Fact{sysinfo.FactCursor},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Cursor", info.Cursor) }},
		item{id: "terminal_font", label: "Term Font", facts: []sysinfo.Fact{sysinfo.FactTerminalFont},
			rows: func(info sysinfo.Info, _ itemEnv) []row { return text("Term Font", info.TerminalFont) }},
		item{id: "media", label: "Media", facts: []sysinfo.Fact{sysinfo.FactMedia}, rows: mediaRows},
	)
}

// stringer returns the rows of an item whose value is formatted already.
// s is used only when state is Present.
func stringer(label string, state sysinfo.State, s string) []row {
	return text(label, sysinfo.Value[string]{State: state, V: s})
}

func localIPRows(info sysinfo.Info, _ itemEnv) []row {
	v := info.LocalIP
	rows := stringer("Local IP", v.State, v.V.Addr.String())
	if v.State == sysinfo.Present {
		rows[0].value = append(rows[0].value, segment{text: "  " + v.V.Interface, role: roleMuted, detail: true})
	}
	return rows
}

// networkRows shows the wireless network by name, or the interface of
// the default route, with the type as a detail.
func networkRows(info sysinfo.Info, _ itemEnv) []row {
	v := info.Network
	name := v.V.Interface
	detail := v.V.Type
	if v.V.SSID != "" {
		name = v.V.SSID
		detail = v.V.Type + ", " + v.V.Interface
	}
	rows := stringer("Network", v.State, name)
	if v.State == sysinfo.Present {
		rows[0].value = append(rows[0].value, segment{text: "  " + detail, role: roleMuted, detail: true})
	}
	return rows
}

func bluetoothRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.Bluetooth.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("Bluetooth")}
	}
	if info.Bluetooth.V {
		return []row{{label: "Bluetooth", badge: &badge{text: "on", look: badgeGood}}}
	}
	return []row{{label: "Bluetooth", badge: &badge{text: "off"}}}
}

// temperature returns a reading colored by its level.
func temperature(c float64) []segment {
	r := roleGood
	switch {
	case c >= tempDanger:
		r = roleDanger
	case c >= tempWarning:
		r = roleWarning
	}
	return []segment{{text: fmt.Sprintf("%.0f°C", c), role: r}}
}

func cpuTempRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.CPUTemp.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("CPU Temp")}
	}
	return []row{{label: "CPU Temp", value: temperature(info.CPUTemp.V)}}
}

func gpuTempRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.GPUTemp.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("GPU Temp")}
	}
	rows := make([]row, len(info.GPUTemp.V))
	for i, c := range info.GPUTemp.V {
		rows[i] = row{label: "GPU Temp", value: temperature(c)}
	}
	return rows
}

func cpuUsageRows(info sysinfo.Info, _ itemEnv) []row {
	if info.CPUUsage.State != sysinfo.Present {
		return []row{unknownRow("CPU Usage")}
	}
	return []row{{label: "CPU Usage", meter: &meter{percent: clampPercent(info.CPUUsage.V)}}}
}

func loadRows(info sysinfo.Info, _ itemEnv) []row {
	l := info.Load.V
	return stringer("Load", info.Load.State, fmt.Sprintf("%.2f %.2f %.2f", l[0], l[1], l[2]))
}

func gpuDriverRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.GPUDriver.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("GPU Driver")}
	}
	rows := make([]row, len(info.GPUDriver.V))
	for i, d := range info.GPUDriver.V {
		rows[i] = row{label: "GPU Driver", value: textValue(d)}
	}
	return rows
}

func gpuMemoryRows(info sysinfo.Info, _ itemEnv) []row {
	switch info.GPUMemory.State {
	case sysinfo.Absent:
		return nil
	case sysinfo.Unknown:
		return []row{unknownRow("VRAM")}
	}
	rows := make([]row, len(info.GPUMemory.V))
	for i, u := range info.GPUMemory.V {
		rows[i] = row{label: "VRAM", meter: sizeMeter(u, u.Total)}
	}
	return rows
}

// biosRows shows the vendor and version with the date as a detail.
func biosRows(info sysinfo.Info, _ itemEnv) []row {
	b := info.BIOS.V
	rows := stringer("BIOS", info.BIOS.State, strings.TrimSpace(b.Vendor+" "+b.Version))
	if info.BIOS.State == sysinfo.Present && b.Date != "" {
		rows[0].value = append(rows[0].value, segment{text: "  " + b.Date, role: roleMuted, detail: true})
	}
	return rows
}

// mediaRows shows "artist - title" with the player state as a badge.
func mediaRows(info sysinfo.Info, _ itemEnv) []row {
	m := info.Media.V
	track := m.Title
	if m.Artist != "" {
		track = m.Artist + " - " + m.Title
	}
	rows := stringer("Media", info.Media.State, track)
	if info.Media.State == sysinfo.Present && m.Status != "" {
		rows[0].badge = &badge{text: strings.ToLower(m.Status)}
	}
	return rows
}

package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// describe writes a row as text for comparisons: muted segments in
// brackets, details in braces, the meter as <percent> with sizes and the
// badge in parentheses.
func describe(r row) string {
	var b strings.Builder
	b.WriteString(r.label + ":")
	if m := r.meter; m != nil {
		fmt.Fprintf(&b, " <%d%%", m.percent)
		if m.sizes {
			fmt.Fprintf(&b, " %d/%d", m.used, m.total)
		}
		if m.inverted {
			b.WriteString(" inverted")
		}
		b.WriteString(">")
	}
	if len(r.value) > 0 {
		b.WriteString(" ")
	}
	for _, s := range r.value {
		switch {
		case s.detail:
			b.WriteString("{" + s.text + "}")
		case s.role == roleMuted:
			b.WriteString("[" + s.text + "]")
		default:
			b.WriteString(s.text)
		}
	}
	if r.badge != nil {
		fmt.Fprintf(&b, " (%s/%d)", r.badge.text, r.badge.look)
	}
	return b.String()
}

// rowsOf builds the rows of item id and describes them.
func rowsOf(id string, info sysinfo.Info, cfg config.Config) []string {
	var out []string
	for _, r := range registry[id].rows(info, itemEnv{cfg: cfg}) {
		out = append(out, describe(r))
	}
	return out
}

func present[T any](v T) sysinfo.Value[T] { return sysinfo.Value[T]{State: sysinfo.Present, V: v} }
func absent[T any]() sysinfo.Value[T]     { return sysinfo.Value[T]{State: sysinfo.Absent} }

func TestItemRows(t *testing.T) {
	const gib = 1 << 30
	showType := config.Defaults()
	showType.Items["memory"] = config.Item{ShowType: true}
	tests := []struct {
		name string
		id   string
		info sysinfo.Info
		cfg  config.Config
		want []string
	}{
		{"os pretty name", "os", sysinfo.Info{OS: present([]byte("ID=ubuntu\nPRETTY_NAME=\"Ubuntu 24.04 LTS\"\n"))}, config.Config{}, []string{"OS: Ubuntu 24.04 LTS"}},
		{"os name and version", "os", sysinfo.Info{OS: present([]byte("NAME=Foo\nVERSION_ID=2\n"))}, config.Config{}, []string{"OS: Foo 2"}},
		{"os arch", "os", sysinfo.Info{OS: present([]byte("ID=arch\nPRETTY_NAME=\"Arch Linux\"\n"))}, config.Config{}, []string{"OS: Arch Linux[ btw]"}},
		{"os empty", "os", sysinfo.Info{OS: present([]byte("#\n"))}, config.Config{}, []string{"OS: [unknown]"}},
		{"os unknown", "os", sysinfo.Info{}, config.Config{}, []string{"OS: [unknown]"}},
		{"kernel", "kernel", sysinfo.Info{Kernel: present("6.8.0")}, config.Config{}, []string{"Kernel: 6.8.0"}},
		{"kernel unknown", "kernel", sysinfo.Info{}, config.Config{}, []string{"Kernel: [unknown]"}},
		{"uptime", "uptime", sysinfo.Info{Uptime: present(3*time.Hour + 25*time.Minute + 7*time.Second)}, config.Config{}, []string{"Uptime: 3h 25m"}},
		{"packages", "packages", sysinfo.Info{Packages: present([]sysinfo.PackageCount{{Manager: "dpkg", Count: 1842}, {Manager: "flatpak", Count: 12}})}, config.Config{},
			[]string{"Packages: 1842{ (dpkg)}, 12{ (flatpak)}"}},
		{"packages absent", "packages", sysinfo.Info{Packages: absent[[]sysinfo.PackageCount]()}, config.Config{}, nil},
		{"terminal absent", "terminal", sysinfo.Info{Terminal: absent[string]()}, config.Config{}, nil},
		{"de", "de", sysinfo.Info{DE: present("GNOME"), WM: present("Mutter")}, config.Config{}, []string{"DE: GNOME"}},
		{"de equal to wm", "de", sysinfo.Info{DE: present("Hyprland"), WM: present("hyprland")}, config.Config{}, nil},
		{"wm with session", "wm", sysinfo.Info{WM: present("Mutter"), Session: present(sysinfo.Wayland)}, config.Config{}, []string{"WM: Mutter (Wayland/0)"}},
		{"wm absent", "wm", sysinfo.Info{WM: absent[string]()}, config.Config{}, nil},
		{"cpu", "cpu", sysinfo.Info{CPU: present(sysinfo.CPU{Model: "AMD Ryzen 7 5800X 8-Core Processor", Count: 16})}, config.Config{},
			[]string{"CPU: AMD Ryzen 7 5800X{  16 cores}"}},
		{"cpu without model", "cpu", sysinfo.Info{CPU: present(sysinfo.CPU{Count: 4})}, config.Config{}, []string{"CPU: [unknown]"}},
		{"gpus", "gpu", sysinfo.Info{GPU: present([]string{"Intel Corporation Alder Lake-P GT2 [Iris Xe Graphics]", "NVIDIA Corporation GA107M [GeForce RTX 3050 Mobile]"})}, config.Config{},
			[]string{"GPU: Intel Iris Xe Graphics", "GPU: NVIDIA GeForce RTX 3050 Mobile"}},
		{"gpu absent", "gpu", sysinfo.Info{GPU: absent[[]string]()}, config.Config{}, nil},
		{"displays", "display", sysinfo.Info{Display: present([]sysinfo.Display{{Width: 2560, Height: 1440}, {Width: 1920, Height: 1080}})}, config.Config{}, []string{"Display: 2560x1440, 1920x1080"}},
		{"battery", "battery", sysinfo.Info{Battery: present([]sysinfo.Battery{{Percent: 78, Status: "Charging"}, {Percent: 15, Status: "Discharging"}, {Percent: 50, Status: "Not charging"}})}, config.Config{},
			[]string{"Battery: <78% inverted> (charging/1)", "Battery: <15% inverted> (discharging/2)", "Battery: <50% inverted> (not charging/0)"}},
		{"battery absent", "battery", sysinfo.Info{Battery: absent[[]sysinfo.Battery]()}, config.Config{}, nil},
		{"battery unknown", "battery", sysinfo.Info{}, config.Config{}, []string{"Battery: [unknown]"}},
		{"memory", "memory", sysinfo.Info{Memory: present(sysinfo.Usage{Used: 8 * gib, Total: 32 * gib})}, config.Defaults(),
			[]string{"RAM: <25% 8589934592/34359738368>"}},
		{"memory types", "memory", sysinfo.Info{Memory: present(sysinfo.Usage{Used: 8 * gib, Total: 32 * gib}), MemoryType: present([]string{"DDR4"})}, showType,
			[]string{"RAM: <25% 8589934592/34359738368> {DDR4}"}},
		{"swap disabled", "swap", sysinfo.Info{Swap: absent[sysinfo.Usage]()}, config.Config{}, []string{"Swap: (disabled/0)"}},
		{"disks", "disks", sysinfo.Info{Disks: present([]sysinfo.Disk{{Mountpoint: "/", Filesystem: "ext4", Usage: sysinfo.Usage{Used: 100, Total: 500, Available: 376}}})}, config.Config{},
			[]string{"/: <21% 100/500> {ext4}"}},
		{"disks unknown", "disks", sysinfo.Info{}, config.Config{}, []string{"Disk: [unknown]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rowsOf(tt.id, tt.info, tt.cfg)
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{42 * time.Second, "42s"},
		{5*time.Minute + 12*time.Second, "5m 12s"},
		{5 * time.Minute, "5m"},
		{3*time.Hour + 25*time.Minute + 7*time.Second, "3h 25m"},
		{3 * time.Hour, "3h"},
		{51*time.Hour + 25*time.Minute, "2d 3h 25m"},
		{48*time.Hour + 30*time.Second, "2d"},
		{48*time.Hour + 5*time.Minute, "2d 5m"},
	}
	for _, tt := range tests {
		if got := formatUptime(tt.d); got != tt.want {
			t.Errorf("formatUptime(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestPercentOf(t *testing.T) {
	tests := []struct {
		part, whole uint64
		want        int
	}{{0, 0, 0}, {1, 3, 33}, {2, 3, 67}, {5, 4, 100}, {0, 10, 0}}
	for _, tt := range tests {
		if got := percentOf(tt.part, tt.whole); got != tt.want {
			t.Errorf("percentOf(%d, %d) = %d, want %d", tt.part, tt.whole, got, tt.want)
		}
	}
	if clampPercent(-5) != 0 || clampPercent(250) != 100 {
		t.Error("clampPercent does not clamp")
	}
}

func TestCleanNames(t *testing.T) {
	cpus := map[string]string{
		"AMD Ryzen 7 5800X 8-Core Processor":        "AMD Ryzen 7 5800X",
		"Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz":   "Intel Core i7-8700",
		"Intel(R) Xeon(R) CPU E3-1230 V2 @ 3.30GHz": "Intel Xeon E3-1230 V2",
		"12th Gen Intel(R) Core(TM) i7-12700H":      "12th Gen Intel Core i7-12700H",
		"AMD Ryzen 9 7950X 16-Core Processor":       "AMD Ryzen 9 7950X",
		"BCM2835":                                   "BCM2835",
		"ARMv7 Processor rev 3 (v7l)":               "ARMv7 rev 3 (v7l)",
	}
	for in, want := range cpus {
		if got := cleanCPU(in); got != want {
			t.Errorf("cleanCPU(%q) = %q, want %q", in, got, want)
		}
	}
	gpus := map[string]string{
		"NVIDIA Corporation GA104 [GeForce RTX 3070]":                                       "NVIDIA GeForce RTX 3070",
		"Advanced Micro Devices, Inc. [AMD/ATI] Navi 21 [Radeon RX 6800/6800 XT / 6900 XT]": "AMD Radeon RX 6800/6800 XT / 6900 XT",
		"Intel Corporation UHD Graphics 620 (rev 07)":                                       "Intel UHD Graphics 620",
		"Matrox Electronics Systems Ltd. G200eR2":                                           "Matrox Electronics Systems Ltd. G200eR2",
		"NVIDIA 2484": "NVIDIA 2484",
		"Intel Corporation Iris Xe Graphics [Intel Iris Xe]": "Intel Iris Xe",
	}
	for in, want := range gpus {
		if got := cleanGPU(in); got != want {
			t.Errorf("cleanGPU(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEveryDefaultItemIsRegistered(t *testing.T) {
	for _, g := range config.Defaults().Groups {
		for _, id := range g.Items {
			if _, ok := registry[id]; !ok {
				t.Errorf("default item %q is not registered", id)
			}
		}
	}
}

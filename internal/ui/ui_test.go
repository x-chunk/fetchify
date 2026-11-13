package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/x-chunk/fetchify/internal/render"
	"github.com/x-chunk/fetchify/internal/sysinfo"
)

const gib = 1 << 30

// sample returns a fully populated snapshot of an ordinary desktop.
func sample() sysinfo.Info {
	return sysinfo.Info{
		User:     "alice",
		Hostname: "box",
		OS:       sysinfo.OS{Name: "ubuntu", Version: "24.04"},
		Kernel:   "6.8.0-45-generic",
		Uptime:   3*time.Hour + 25*time.Minute + 7*time.Second,
		Shell:    "zsh",
		Desktop:  sysinfo.Desktop{WM: "i3", Server: sysinfo.X11},
		CPU:      sysinfo.CPU{Model: "AMD Ryzen 7 5800X", Cores: 16},
		GPUs:     []string{"NVIDIA GeForce RTX 3070"},
		Memory: &sysinfo.Memory{
			Usage: sysinfo.Usage{Used: 8 * gib, Total: 32 * gib, Percent: 25},
		},
		Swap: &sysinfo.Usage{Used: gib, Total: 4 * gib, Percent: 25},
		Disks: []sysinfo.Disk{
			{Usage: sysinfo.Usage{Used: 100e9, Total: 500e9, Percent: 21}, Mountpoint: "/", Filesystem: "ext4"},
			{Usage: sysinfo.Usage{Used: 250e9, Total: 1000e9, Percent: 25}, Mountpoint: "/home", Filesystem: "btrfs"},
		},
	}
}

// plain renders b without styles and strips trailing spaces, which are
// padding and not part of what the test pins down.
func plain(b render.Block) string {
	lines := strings.Split(render.NewEngine(render.ProfilePlain).String(b), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func TestBuild(t *testing.T) {
	want := strings.Join([]string{
		`==== Fetchify ====   alice@box`,
		`   __         __     ─────────`,
		` /"  "\     /"  "\   OS: ubuntu 24.04`,
		`(  (\  )___(  /)  )  Kernel: 6.8.0-45-generic`,
		` \               /   Uptime: 3h25m7s`,
		` /               \   Shell: zsh`,
		`/    () ___ ()    \  WM: i3 (X11)`,
		`|      (   )      |  CPU: AMD Ryzen 7 5800X (16 cores)`,
		` \      \_/      /   GPU: NVIDIA GeForce RTX 3070`,
		`   \...__!__.../     RAM: 8.00 GiB / 32.00 Gib (25%)`,
		`        " mh         Swap: 1.00 GiB / 4.00 Gib (25%)`,
		`   v1.0.4-generic    Disk (/): 100.00 GiB / 500.00 GiB (21%) - ext4`,
		`                     Disk (/home): 250.00 GiB / 1000.00 GiB (25%) - btrfs`,
		``,
		`                     ████████████████████████`,
		`                     ████████████████████████`,
		``,
	}, "\n")
	if got := plain(Build(sample())); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestBuildWithNothingKnown(t *testing.T) {
	want := strings.Join([]string{
		`==== Fetchify ====   @`,
		`   __         __     ─`,
		` /"  "\     /"  "\   OS: Unknown`,
		`(  (\  )___(  /)  )  Kernel: Unknown`,
		` \               /   Uptime: 0s`,
		` /               \   Shell: Unknown`,
		`/    () ___ ()    \  WM: Unknown`,
		`|      (   )      |  CPU: Unknown (0 cores)`,
		` \      \_/      /   GPU: Unknown`,
		`   \...__!__.../     RAM: Error`,
		`        " mh         Swap: Error`,
		`   v1.0.4-generic`,
		`                     ████████████████████████`,
		`                     ████████████████████████`,
		``,
	}, "\n")
	if got := plain(Build(sysinfo.Info{})); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestHeader(t *testing.T) {
	ansi := render.NewEngine(render.ProfileANSI16)

	user := sysinfo.Info{User: "alice", Hostname: "box"}
	want := "\x1b[0;31malice\x1b[0;32m@\x1b[0;36mbox\x1b[0m\n" +
		"\x1b[0;32m─────────\x1b[0m"
	if got := ansi.String(header(user)); got != want {
		t.Errorf("user header = %q, want %q", got, want)
	}

	root := sysinfo.Info{User: "root", Hostname: "box", Root: true}
	block := header(root)
	if got := block[0].String(); got != "root@box - Don't login as root ^-^" {
		t.Errorf("root title = %q", got)
	}
	if block[1].Width() != block[0].Width() {
		t.Errorf("rule width = %d, want the title width %d", block[1].Width(), block[0].Width())
	}
	if got := ansi.String(block[1:]); !strings.HasPrefix(got, "\x1b[0;31m") {
		t.Errorf("root rule = %q, want the warning style", got)
	}
}

func TestSections(t *testing.T) {
	tests := []struct {
		name    string
		section section
		edit    func(*sysinfo.Info)
		want    []string
	}{
		{"arch", osSection, func(i *sysinfo.Info) { i.OS = sysinfo.OS{Name: "arch"} }, []string{"OS: i use arch, btw"}},
		{"wayland", desktopSection, func(i *sysinfo.Info) { i.Desktop.Server = sysinfo.Wayland }, []string{"WM: i3 (Wayland)"}},
		{
			"several gpus", gpuSection,
			func(i *sysinfo.Info) { i.GPUs = []string{"Intel UHD", "NVIDIA RTX"} },
			[]string{"GPU: Intel UHD", "GPU: NVIDIA RTX"},
		},
		{
			"memory types", memorySection,
			func(i *sysinfo.Info) { i.Memory.Types = []string{"DDR4", "DDR5"} },
			[]string{"RAM: 8.00 GiB / 32.00 Gib (25%) - DDR4, DDR5"},
		},
		{"swap disabled", swapSection, func(i *sysinfo.Info) { i.Swap = &sysinfo.Usage{} }, []string{"Swap: Disabled"}},
		{"no disks", diskSection, func(i *sysinfo.Info) { i.Disks = nil }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := sample()
			tt.edit(&info)
			var got []string
			for _, line := range keyValue(tt.section(info)) {
				got = append(got, line.String())
			}
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSanitizesCollectedText(t *testing.T) {
	info := sample()
	info.Hostname = "box\x1b[2J"
	info.GPUs = []string{"evil\ncard"}
	out := render.NewEngine(render.ProfilePlain).String(Build(info))
	if strings.ContainsRune(out, '\x1b') {
		t.Error("output contains an escape character from collected text")
	}
	if got, want := strings.Count(out, "\n"), Build(sample()).Height()-1; got != want {
		t.Errorf("output has %d newlines, want %d", got, want)
	}
}

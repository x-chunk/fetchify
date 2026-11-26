package ui

import (
	"strings"
	"testing"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/sysinfo"
)

const gib = 1 << 30

// sample returns a fully populated snapshot of an ordinary desktop.
func sample() sysinfo.Info {
	return sysinfo.Info{
		User:     sysinfo.Value[sysinfo.User]{State: sysinfo.Present, V: sysinfo.User{Name: "alice", UID: 1000}},
		Hostname: sysinfo.Value[string]{State: sysinfo.Present, V: "box"},
		OS:       sysinfo.Value[[]byte]{State: sysinfo.Present, V: []byte("ID=ubuntu\nVERSION_ID=24.04\n")},
		Kernel:   sysinfo.Value[string]{State: sysinfo.Present, V: "6.8.0-45-generic"},
		Uptime:   sysinfo.Value[time.Duration]{State: sysinfo.Present, V: 3*time.Hour + 25*time.Minute + 7*time.Second},
		Shell:    sysinfo.Value[string]{State: sysinfo.Present, V: "zsh"},
		Desktop:  sysinfo.Desktop{WM: "i3", Server: sysinfo.X11},
		CPU:      sysinfo.CPU{Model: "AMD Ryzen 7 5800X", Cores: 16},
		GPUs:     []string{"NVIDIA GeForce RTX 3070"},
		Memory: &sysinfo.Memory{
			Usage: sysinfo.Usage{Used: 8 * gib, Total: 32 * gib, Percent: 25},
		},
		Swap: &sysinfo.Usage{Used: gib, Total: 4 * gib, Percent: 25},
		Disks: []sysinfo.Disk{
			{Usage: sysinfo.Usage{Used: 100 * gib, Total: 500 * gib, Percent: 21}, Mountpoint: "/", Filesystem: "ext4"},
			{Usage: sysinfo.Usage{Used: 250 * gib, Total: 1000 * gib, Percent: 25}, Mountpoint: "/home", Filesystem: "btrfs"},
		},
	}
}

// render renders c in an unbounded context at profile p.
func render(t *testing.T, c beautify.Composable, p dra.Profile) []byte {
	t.Helper()
	m, err := c.Render(beautify.NewContext(nil))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return dra.NewEngine(p).Append(nil, m)
}

// plain renders c without styles and strips trailing spaces, which are
// padding and not part of what the test pins down.
func plain(t *testing.T, c beautify.Composable) string {
	t.Helper()
	lines := strings.Split(string(render(t, c, dra.ProfilePlain)), "\n")
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
		`   \...__!__.../     RAM: 8.00 GiB / 32.00 GiB (25%)`,
		`        " mh         Swap: 1.00 GiB / 4.00 GiB (25%)`,
		`   v1.0.4-generic    Disk (/): 100.00 GiB / 500.00 GiB (21%) - ext4`,
		`                     Disk (/home): 250.00 GiB / 1000.00 GiB (25%) - btrfs`,
		``,
		`                     ████████████████████████`,
		`                     ████████████████████████`,
		``,
	}, "\n")
	if got := plain(t, Build(sample())); got != want {
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
	if got := plain(t, Build(sysinfo.Info{})); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestHeader(t *testing.T) {
	user := sample()
	want := "\x1b[0;31malice\x1b[0;32m@\x1b[0;36mbox\x1b[0m\n" +
		"\x1b[0;32m─────────\x1b[0m"
	if got := string(render(t, header(user), dra.ProfileANSI16)); got != want {
		t.Errorf("user header = %q, want %q", got, want)
	}

	root := sample()
	root.User = sysinfo.Value[sysinfo.User]{State: sysinfo.Present, V: sysinfo.User{Name: "root"}}
	lines := strings.Split(string(render(t, header(root), dra.ProfilePlain)), "\n")
	if len(lines) != 2 || lines[0] != "root@box - Don't login as root ^-^" {
		t.Fatalf("root header = %q", lines)
	}
	if got, want := len([]rune(lines[1])), len([]rune(lines[0])); got != want {
		t.Errorf("rule width = %d, want the title width %d", got, want)
	}
	ansi := strings.Split(string(render(t, header(root), dra.ProfileANSI16)), "\n")
	if !strings.HasPrefix(ansi[1], "\x1b[0;31m") {
		t.Errorf("root rule = %q, want the warning style", ansi[1])
	}
}

func TestSections(t *testing.T) {
	tests := []struct {
		name    string
		section section
		edit    func(*sysinfo.Info)
		want    []string
	}{
		{"arch", osSection, func(i *sysinfo.Info) { i.OS.V = []byte("ID=arch\nBUILD_ID=rolling\n") }, []string{"OS: i use arch, btw"}},
		{"wayland", desktopSection, func(i *sysinfo.Info) { i.Desktop.Server = sysinfo.Wayland }, []string{"WM: i3 (Wayland)"}},
		{
			"several gpus", gpuSection,
			func(i *sysinfo.Info) { i.GPUs = []string{"Intel UHD", "NVIDIA RTX"} },
			[]string{"GPU: Intel UHD", "GPU: NVIDIA RTX"},
		},
		{
			"memory types", memorySection,
			func(i *sysinfo.Info) { i.Memory.Types = []string{"DDR4", "DDR5"} },
			[]string{"RAM: 8.00 GiB / 32.00 GiB (25%) - DDR4, DDR5"},
		},
		{"swap disabled", swapSection, func(i *sysinfo.Info) { i.Swap = &sysinfo.Usage{} }, []string{"Swap: Disabled"}},
		{"no disks", diskSection, func(i *sysinfo.Info) { i.Disks = nil }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := sample()
			tt.edit(&info)
			var got []string
			for _, e := range tt.section(info, dra.NewFactory()) {
				got = append(got, e.key+": "+e.value.String())
			}
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSanitizesCollectedText(t *testing.T) {
	info := sample()
	info.Hostname.V = "box\x1b[2J"
	info.GPUs = []string{"evil\ncard"}
	out := string(render(t, Build(info), dra.ProfilePlain))
	if strings.ContainsRune(out, '\x1b') {
		t.Error("output contains an escape character from collected text")
	}
	clean := string(render(t, Build(sample()), dra.ProfilePlain))
	if got, want := strings.Count(out, "\n"), strings.Count(clean, "\n"); got != want {
		t.Errorf("output has %d newlines, want %d", got, want)
	}
}

func TestFactsAreKnown(t *testing.T) {
	known := map[sysinfo.Fact]bool{}
	for _, f := range sysinfo.Facts() {
		known[f] = true
	}
	for _, f := range Facts() {
		if !known[f] {
			t.Errorf("ui needs %q, which sysinfo does not know", f)
		}
	}
}

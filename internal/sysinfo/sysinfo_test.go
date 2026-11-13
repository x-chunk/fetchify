package sysinfo

import (
	"errors"
	"slices"
	"testing"
)

// env returns a getenv function backed by vars.
func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// commands returns a runner that answers with the canned output for the
// command name and fails for every other command.
func commands(outputs map[string]string) runner {
	return func(name string, _ ...string) ([]byte, error) {
		out, ok := outputs[name]
		if !ok {
			return nil, errors.New("command not found")
		}
		return []byte(out), nil
	}
}

func TestShell(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"/usr/bin/zsh", "zsh"},
		{"bash", "bash"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := shell(env(map[string]string{"SHELL": tt.value})); got != tt.want {
			t.Errorf("shell(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestDesktop(t *testing.T) {
	tests := []struct {
		name     string
		commands map[string]string
		env      map[string]string
		want     Desktop
	}{
		{
			name: "nothing available",
			want: Desktop{Server: X11},
		},
		{
			name:     "xprop wins",
			commands: map[string]string{"xprop": "_NET_WM_NAME(UTF8_STRING) = \"i3\"\n", "wmctrl": "Name: other\n"},
			env:      map[string]string{"XDG_CURRENT_DESKTOP": "GNOME"},
			want:     Desktop{WM: "i3", Server: X11},
		},
		{
			name:     "wmctrl when xprop has no value",
			commands: map[string]string{"xprop": "_NET_WM_NAME:  not found.\n", "wmctrl": "Name: GNOME Shell\nClass: N/A\n"},
			want:     Desktop{WM: "GNOME Shell", Server: X11},
		},
		{
			name: "environment as the last resort",
			env:  map[string]string{"XDG_CURRENT_DESKTOP": "KDE", "WAYLAND_DISPLAY": "wayland-0"},
			want: Desktop{WM: "KDE", Server: Wayland},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := desktop(commands(tt.commands), env(tt.env)); got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseLspci(t *testing.T) {
	out := "00:00.0 Host bridge: Intel Corporation Device 9b61\n" +
		"00:02.0 VGA compatible controller: Intel Corporation UHD Graphics (rev 02)\n" +
		"01:00.0 VGA compatible controller: NVIDIA Corporation Device 1f99: Max-Q\n" +
		"02:00.0 VGA compatible controller\n"
	want := []string{
		"Intel Corporation UHD Graphics (rev 02)",
		"NVIDIA Corporation Device 1f99: Max-Q",
	}
	if got := parseLspci(out); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := parseLspci(""); got != nil {
		t.Errorf("parseLspci(\"\") = %q, want nil", got)
	}
}

func TestGPUsWithoutLspci(t *testing.T) {
	if got := gpus(commands(nil)); got != nil {
		t.Errorf("got %q, want nil", got)
	}
}

func TestParseMemoryTypes(t *testing.T) {
	out := "Physical Memory Array\n" +
		"\tUse: System Memory\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"\tType Detail: Synchronous\n" +
		"Memory Device\n" +
		"\tType: Unknown\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"Memory Device\n" +
		"\tType: Other\n" +
		"Memory Device\n" +
		"\tType: RAM\n" +
		"Memory Device\n" +
		"\tType: DDR5\n"
	want := []string{"DDR4", "DDR5"}
	if got := parseMemoryTypes(out); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIsIgnoredMount(t *testing.T) {
	tests := []struct {
		mount string
		want  bool
	}{
		{"/", false},
		{"/home", false},
		{"/boot", true},
		{"/boot/efi", true},
		{"/var/lib/docker", true},
		{"/variable", false},
		{"/opt/run", false},
	}
	for _, tt := range tests {
		if got := isIgnoredMount(tt.mount); got != tt.want {
			t.Errorf("isIgnoredMount(%q) = %v, want %v", tt.mount, got, tt.want)
		}
	}
}

func TestCollect(t *testing.T) {
	info := Collect()
	if info.CPU.Cores < 1 {
		t.Errorf("CPU.Cores = %d, want at least 1", info.CPU.Cores)
	}
	if info.Desktop.Server != X11 && info.Desktop.Server != Wayland {
		t.Errorf("Desktop.Server = %q", info.Desktop.Server)
	}
}

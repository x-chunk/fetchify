package sysinfo

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestParseXprop(t *testing.T) {
	if got := parseXpropWindow("_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x1a00003\n"); got != "0x1a00003" {
		t.Errorf("window = %q", got)
	}
	if got := parseXpropWindow("_NET_SUPPORTING_WM_CHECK:  not found.\n"); got != "" {
		t.Errorf("window without property = %q", got)
	}
	if got := parseXprop("_NET_WM_NAME(UTF8_STRING) = \"i3\"\n"); got != "i3" {
		t.Errorf("name = %q", got)
	}
	if got := parseXprop("_NET_WM_NAME:  not found.\n"); got != "" {
		t.Errorf("name without property = %q", got)
	}
	if got := parseWmctrl("Name: GNOME Shell\nClass: N/A\n"); got != "GNOME Shell" {
		t.Errorf("wmctrl = %q", got)
	}
	if got := parseWmctrl("garbage"); got != "" {
		t.Errorf("wmctrl garbage = %q", got)
	}
}

func TestWM(t *testing.T) {
	procs := fstest.MapFS{
		"proc/10/comm":    {Data: []byte("systemd\n")},
		"proc/20/comm":    {Data: []byte("xmonad-x86_64-l\n")},
		"proc/self/comm":  {Data: []byte("i3\n")}, // not numeric: ignored
		"proc/30/comm":    {Data: []byte("bash\n")},
		"proc/uptime":     {Data: []byte("1 1\n")},
		"proc/40/missing": {},
	}
	x11 := map[string]string{"DISPLAY": ":0"}
	tests := []struct {
		name string
		src  Source
		want Value[string]
	}{
		{"wayland compositor", ubuntuDesktop.source(), present("Mutter")},
		{"kwin", btrfsBox.source(), present("KWin")},
		{"xprop", archLaptop.source(), present("i3")},
		{"no session", sshServer.source(), absent[string]()},
		{"wmctrl", machine{env: x11, commands: map[string]string{"wmctrl -m": "Name: Openbox\n"}}.source(), present("Openbox")},
		{"truncated process name", Source{FS: procs, Getenv: func(k string) string { return x11[k] }}, present("xmonad")},
		{"nothing found", machine{env: x11}.source(), Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactWM}, Options{}).WM
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

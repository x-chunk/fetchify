package sysinfo

import "testing"

func TestDE(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Value[string]
	}{
		{"first entry", map[string]string{"XDG_SESSION_TYPE": "wayland", "XDG_CURRENT_DESKTOP": "ubuntu:GNOME"}, present("ubuntu")},
		{"desktop session", map[string]string{"DISPLAY": ":0", "DESKTOP_SESSION": "xfce"}, present("xfce")},
		{"bare window manager", map[string]string{"DISPLAY": ":0"}, absent[string]()},
		{"no session", map[string]string{"XDG_CURRENT_DESKTOP": "KDE"}, absent[string]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, machine{env: tt.env}, FactDE).DE; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

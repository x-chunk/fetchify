package sysinfo

import "testing"

func TestSession(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Value[SessionType]
	}{
		{"xdg wayland", map[string]string{"XDG_SESSION_TYPE": "wayland", "DISPLAY": ":0"}, present(Wayland)},
		{"xdg x11", map[string]string{"XDG_SESSION_TYPE": "x11"}, present(X11)},
		{"startx from a tty", map[string]string{"XDG_SESSION_TYPE": "tty", "DISPLAY": ":0"}, present(X11)},
		{"wayland display", map[string]string{"WAYLAND_DISPLAY": "wayland-0", "DISPLAY": ":0"}, present(Wayland)},
		{"console", map[string]string{"XDG_SESSION_TYPE": "tty"}, absent[SessionType]()},
		{"nothing", nil, absent[SessionType]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, machine{env: tt.env}, FactSession).Session; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

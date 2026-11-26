package sysinfo

import (
	"reflect"
	"testing"
)

func TestDisplay(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[[]Display]
	}{
		{"two outputs", archLaptop, present([]Display{{1920, 1200}, {2560, 1440}})},
		{"one output", ubuntuDesktop, present([]Display{{2560, 1440}})},
		{"headless", sshServer, absent[[]Display]()},
		{"no drm", container, Value[[]Display]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactDisplay).Display; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		in   string
		want Display
		ok   bool
	}{
		{"2560x1440", Display{2560, 1440}, true},
		{"1920x1080i", Display{1920, 1080}, true},
		{"", Display{}, false},
		{"axb", Display{}, false},
		{"0x0", Display{}, false},
	}
	for _, tt := range tests {
		if got, ok := parseMode(tt.in); got != tt.want || ok != tt.ok {
			t.Errorf("parseMode(%q) = %v, %v", tt.in, got, ok)
		}
	}
}

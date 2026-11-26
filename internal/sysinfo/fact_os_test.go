package sysinfo

import (
	"strings"
	"testing"
)

func TestOS(t *testing.T) {
	tests := []struct {
		machine machine
		state   State
		id      string
	}{
		{ubuntuDesktop, Present, "ID=ubuntu"},
		{container, Present, "ID=alpine"},
		{placeholderDMI, Present, "ID=linuxmint"}, // only usr/lib/os-release
		{machine{}, Unknown, ""},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			got := collectOne(t, tt.machine, FactOS).OS
			if got.State != tt.state || !strings.Contains(string(got.V), tt.id) {
				t.Errorf("got %v %q, want %v with %q", got.State, got.V, tt.state, tt.id)
			}
		})
	}
}

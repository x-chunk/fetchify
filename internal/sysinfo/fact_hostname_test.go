package sysinfo

import "testing"

func TestHostname(t *testing.T) {
	tests := []struct {
		machine machine
		want    Value[string]
	}{
		{ubuntuDesktop, present("box")},
		{container, present("3f2a9c1b7d4e")},
		{placeholderDMI, Value[string]{}},
		{machine{}, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactHostname).Hostname; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

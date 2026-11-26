package sysinfo

import "testing"

func TestKernel(t *testing.T) {
	tests := []struct {
		machine machine
		want    Value[string]
	}{
		{ubuntuDesktop, present("6.8.0-45-generic")},
		{raspberryPi, present("6.6.31+rpt-rpi-v8")},
		{machine{}, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactKernel).Kernel; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

package sysinfo

import "testing"

func TestCPU(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[CPU]
	}{
		{"x86", ubuntuDesktop, present(CPU{Model: "AMD Ryzen 7 5800X 8-Core Processor", Count: 4})},
		{"arm without model name", raspberryPi, present(CPU{Model: "BCM2835", Count: 4})},
		{"missing", machine{}, Value[CPU]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactCPU).CPU; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseCPUInfo(t *testing.T) {
	tests := []struct {
		in   string
		want CPU
	}{
		{"processor\t: 0\nModel\t\t: Raspberry Pi 4\n", CPU{Model: "Raspberry Pi 4", Count: 1}},
		{"processor : 0\nmodel name : A\nprocessor : 1\nmodel name : B\n", CPU{Model: "A", Count: 2}},
		{"", CPU{}},
	}
	for _, tt := range tests {
		if got := parseCPUInfo(tt.in); got != tt.want {
			t.Errorf("parseCPUInfo(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

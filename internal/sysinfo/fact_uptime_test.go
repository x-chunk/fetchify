package sysinfo

import (
	"testing"
	"time"
)

func TestParseUptime(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"12307.67 190123.40\n", 12307 * time.Second, true},
		{"0.99 1.20", 0, true},
		{"42", 42 * time.Second, true},
		{"", 0, false},
		{"abc 1", 0, false},
		{"-5 1", 0, false},
		{"1e300 1", 0, false},
	}
	for _, tt := range tests {
		got, ok := parseUptime(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("parseUptime(%q) = %v, %v; want %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestUptime(t *testing.T) {
	tests := []struct {
		machine machine
		want    Value[time.Duration]
	}{
		{archLaptop, present(26*time.Hour + 3*time.Minute + 4*time.Second)},
		{placeholderDMI, present(time.Duration(0))},
		{machine{}, Value[time.Duration]{}},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactUptime).Uptime; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

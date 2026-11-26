package sysinfo

import "testing"

func TestSwap(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[Usage]
	}{
		{"in use", ubuntuDesktop, present(Usage{Used: gib, Total: 4 * gib, Available: 3 * gib})},
		{"disabled", archLaptop, absent[Usage]()},
		{"missing", machine{}, Value[Usage]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactSwap).Swap; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

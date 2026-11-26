package sysinfo

import (
	"reflect"
	"testing"
)

func TestGPUDriver(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[[]string]
	}{
		{"two GPUs", archLaptop, present([]string{"i915", "nvidia"})},
		{"no uevent", raspberryPi, absent[[]string]()},
		{"no GPU", container, absent[[]string]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactGPUDriver).GPUDriver; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGPUMemory(t *testing.T) {
	smi := archLaptop
	smi.commands = map[string]string{"nvidia-smi --query-gpu=memory.used,memory.total --format=csv,noheader,nounits": "512, 4096\n"}
	tests := []struct {
		name    string
		machine machine
		want    Value[[]Usage]
	}{
		{"nvidia-smi", smi, present([]Usage{{Used: 512 * mib, Total: 4096 * mib, Available: 3584 * mib}})},
		{"amdgpu", btrfsBox, present([]Usage{{Used: 2 * gib, Total: 16 * gib, Available: 14 * gib}})},
		{"nothing reports", archLaptop, Value[[]Usage]{}},
		{"no GPU", container, absent[[]Usage]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactGPUMemory).GPUMemory; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

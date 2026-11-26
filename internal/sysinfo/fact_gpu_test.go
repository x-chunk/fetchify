package sysinfo

import (
	"maps"
	"reflect"
	"testing"
)

func TestGPU(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[[]string]
	}{
		{"pci.ids, two GPUs, 3D controller", archLaptop, present([]string{
			"Intel Corporation Alder Lake-P GT2 [Iris Xe Graphics]",
			"NVIDIA Corporation GA107M [GeForce RTX 3050 Mobile]",
		})},
		{"lspci", ubuntuDesktop, present([]string{"NVIDIA Corporation GA104 [GeForce RTX 3070]"})},
		{"other database path", btrfsBox, present([]string{"Advanced Micro Devices, Inc. [AMD/ATI] 744c"})},
		{"no names at all", sshServer, present([]string{"102b 0536"})},
		{"no PCI device", raspberryPi, absent[[]string]()},
		{"no DRM", container, absent[[]string]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactGPU).GPU; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseLspciMM(t *testing.T) {
	out := `00:02.0 "VGA compatible controller" "Intel Corporation" "Alder Lake-P GT2 [Iris Xe Graphics]" -r0c -p00 "Dell" "Device 0b19"` + "\n" +
		`01:00.0 "3D controller" "NVIDIA Corporation" "GA107M [GeForce RTX 3050 Mobile]" -ra1 "Dell" "Device 0b19"` + "\n" +
		"garbage\n"
	want := map[string]string{
		"00:02.0": "Intel Corporation Alder Lake-P GT2 [Iris Xe Graphics]",
		"01:00.0": "NVIDIA Corporation GA107M [GeForce RTX 3050 Mobile]",
	}
	if got := parseLspciMM(out); !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

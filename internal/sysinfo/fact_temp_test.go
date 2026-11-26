package sysinfo

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestCPUTemp(t *testing.T) {
	file := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s + "\n")} }
	tests := []struct {
		name string
		fs   fstest.MapFS
		want Value[float64]
	}{
		{"preferred driver", fstest.MapFS{
			"sys/class/hwmon/hwmon0/name":        file("nvme"),
			"sys/class/hwmon/hwmon0/temp1_input": file("40000"),
			"sys/class/hwmon/hwmon1/name":        file("coretemp"),
			"sys/class/hwmon/hwmon1/temp1_input": file("61000"),
			"sys/class/hwmon/hwmon2/name":        file("k10temp"),
			"sys/class/hwmon/hwmon2/temp1_input": file("54500"),
		}, present(54.5)},
		{"thermal zone", fstest.MapFS{
			"sys/class/thermal/thermal_zone0/type": file("acpitz"),
			"sys/class/thermal/thermal_zone0/temp": file("30000"),
			"sys/class/thermal/thermal_zone1/type": file("cpu-thermal"),
			"sys/class/thermal/thermal_zone1/temp": file("48312"),
		}, present(48.312)},
		{"no sensor", fstest.MapFS{}, absent[float64]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), Source{FS: tt.fs}, []Fact{FactCPUTemp}, Options{}).CPUTemp
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGPUTemp(t *testing.T) {
	smi := archLaptop
	smi.commands = map[string]string{"nvidia-smi --query-gpu=temperature.gpu --format=csv,noheader,nounits": "47\n"}
	amd := btrfsBox
	tests := []struct {
		name    string
		machine machine
		want    Value[[]float64]
	}{
		{"hwmon and nvidia-smi", smi, present([]float64{45, 47})},
		{"nvidia without nvidia-smi", archLaptop, present([]float64{45})},
		{"amd hwmon", amd, present([]float64{38})},
		{"no reading", sshServer, Value[[]float64]{}},
		{"no GPU", container, absent[[]float64]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactGPUTemp).GPUTemp; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

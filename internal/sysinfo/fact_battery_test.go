package sysinfo

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestBattery(t *testing.T) {
	file := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s + "\n")} }
	tests := []struct {
		name string
		src  Source
		want Value[[]Battery]
	}{
		{"laptop", archLaptop.source(), present([]Battery{{Percent: 78, Status: "Charging"}})},
		{"desktop", ubuntuDesktop.source(), absent[[]Battery]()},
		{"no power supply class", container.source(), absent[[]Battery]()},
		{"energy counters", Source{FS: fstest.MapFS{
			"sys/class/power_supply/BAT1/type":        file("Battery"),
			"sys/class/power_supply/BAT1/energy_now":  file("25000000"),
			"sys/class/power_supply/BAT1/energy_full": file("50000000"),
		}}, present([]Battery{{Percent: 50, Status: "Unknown"}})},
		{"unreadable", Source{FS: fstest.MapFS{
			"sys/class/power_supply/BAT0/type": file("Battery"),
		}}, Value[[]Battery]{}},
		{"clamped", Source{FS: fstest.MapFS{
			"sys/class/power_supply/BAT0/type":     file("Battery"),
			"sys/class/power_supply/BAT0/capacity": file("104"),
			"sys/class/power_supply/BAT0/status":   file("Full"),
		}}, present([]Battery{{Percent: 100, Status: "Full"}})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactBattery}, Options{}).Battery
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

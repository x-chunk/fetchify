package sysinfo

import "testing"

func TestBoard(t *testing.T) {
	tests := []struct {
		machine machine
		want    Value[string]
	}{
		{btrfsBox, present("Micro-Star International Co., Ltd. MAG B550 TOMAHAWK (MS-7C91)")},
		{placeholderDMI, present("ASRock")},
		{container, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactBoard).Board; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBIOS(t *testing.T) {
	tests := []struct {
		machine machine
		want    Value[BIOS]
	}{
		{btrfsBox, present(BIOS{Vendor: "American Megatrends International, LLC.", Version: "1.A0", Date: "08/09/2023"})},
		{placeholderDMI, present(BIOS{Version: "P2.10"})},
		{container, Value[BIOS]{}},
	}
	for _, tt := range tests {
		t.Run(tt.machine.dir, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactBIOS).BIOS; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

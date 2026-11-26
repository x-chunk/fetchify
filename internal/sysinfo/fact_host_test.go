package sysinfo

import "testing"

func TestHostModel(t *testing.T) {
	tests := []struct {
		name, version, want string
	}{
		{"XPS 13 9310", "", "XPS 13 9310"},
		{"21HMCTO1WW", "ThinkPad X1 Carbon Gen 11", "ThinkPad X1 Carbon Gen 11"},
		{"MS-7C56", "1.0", "MS-7C56 1.0"},
		{"OptiPlex 7090", "OptiPlex 7090", "OptiPlex 7090"},
		{"System Product Name", "System Version", ""},
		{"PowerEdge R640", "Not Specified", "PowerEdge R640"},
		{"To Be Filled By O.E.M.", "TO BE FILLED BY O.E.M.", ""},
		{"", "", ""},
		{"Default string", "none", ""},
		{"20XW", "ThinkPad", "20XW ThinkPad"}, // too short for a model code
	}
	for _, tt := range tests {
		if got := hostModel(tt.name, tt.version); got != tt.want {
			t.Errorf("hostModel(%q, %q) = %q, want %q", tt.name, tt.version, got, tt.want)
		}
	}
}

func TestHost(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[string]
	}{
		{"dmi", ubuntuDesktop, present("ThinkPad X1 Carbon Gen 11")},
		{"device tree", raspberryPi, present("Raspberry Pi 4 Model B Rev 1.4")},
		{"placeholder version", sshServer, present("PowerEdge R640")},
		{"placeholders", placeholderDMI, Value[string]{}},
		{"container", container, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactHost).Host; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

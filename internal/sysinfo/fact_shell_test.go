package sysinfo

import "testing"

func TestShell(t *testing.T) {
	tests := []struct {
		value string
		want  Value[string]
	}{
		{"/usr/bin/zsh", present("zsh")},
		{"bash", present("bash")},
		{"", Value[string]{}},
	}
	for _, tt := range tests {
		m := machine{env: map[string]string{"SHELL": tt.value}}
		if got := collectOne(t, m, FactShell).Shell; got != tt.want {
			t.Errorf("SHELL=%q: got %+v, want %+v", tt.value, got, tt.want)
		}
	}
}

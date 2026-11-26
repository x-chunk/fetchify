package sysinfo

import "testing"

func TestParseStatPPID(t *testing.T) {
	tests := []struct {
		in   string
		want int
		ok   bool
	}{
		{"777 (fetchify) R 700 777 700 0", 700, true},
		{"5 (a b) (c)) S 4 5", 4, true},
		{"5 (x) S", 0, false},
		{"garbage", 0, false},
		{"5 (x) S abc", 0, false},
	}
	for _, tt := range tests {
		got, err := parseStatPPID(tt.in)
		if got != tt.want || (err == nil) != tt.ok {
			t.Errorf("parseStatPPID(%q) = %d, %v; want %d, ok %v", tt.in, got, err, tt.want, tt.ok)
		}
	}
}

func TestTerminal(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[string]
	}{
		{"variable of the terminal", ubuntuDesktop, present("GNOME Terminal")},
		{"process tree", archLaptop, present("st")},
		{"ssh", sshServer, absent[string]()},
		{"no parent", container, absent[string]()},
		{"linux console", raspberryPi, absent[string]()},
		{"multiplexer", btrfsBox, present("tmux")},
		{"no process table", placeholderDMI, Value[string]{}},
		{"TERM_PROGRAM", machine{env: map[string]string{"TERM_PROGRAM": "iTerm.app"}}, present("iTerm")},
		{"foot", machine{env: map[string]string{"TERM": "foot-extra"}}, present("foot")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactTerminal).Terminal; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

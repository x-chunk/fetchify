package sysinfo

import (
	"reflect"
	"slices"
	"testing"
)

func TestParseMemoryTypes(t *testing.T) {
	out := "Physical Memory Array\n" +
		"\tUse: System Memory\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"\tType Detail: Synchronous\n" +
		"Memory Device\n" +
		"\tType: Unknown\n" +
		"Memory Device\n" +
		"\tType: DDR4\n" +
		"Memory Device\n" +
		"\tType: Other\n" +
		"Memory Device\n" +
		"\tType: RAM\n" +
		"Memory Device\n" +
		"\tType: DDR5\n"
	want := []string{"DDR4", "DDR5"}
	if got := parseMemoryTypes(out); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMemoryType(t *testing.T) {
	root := sshServer
	root.commands = map[string]string{"dmidecode -t memory": "Memory Device\n\tType: DDR4\n"}
	tests := []struct {
		name    string
		machine machine
		want    Value[[]string]
	}{
		{"root", root, present([]string{"DDR4"})},
		{"not root", ubuntuDesktop, absent[[]string]()},
		{"no dmidecode", sshServer, Value[[]string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactMemoryType).MemoryType; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

package sysinfo

import (
	"context"
	"encoding/binary"
	"testing"
	"testing/fstest"
)

// utmpRecord builds a utmp record of type kind for user.
func utmpRecord(kind uint16, user string) []byte {
	r := make([]byte, utmpSize)
	binary.LittleEndian.PutUint16(r, kind)
	copy(r[utmpUserOffset:], user)
	return r
}

func TestInitAndProcesses(t *testing.T) {
	info := Collect(context.Background(), ubuntuDesktop.source(), []Fact{FactInit, FactProcesses}, Options{})
	if info.Init != present("systemd") {
		t.Errorf("init = %+v", info.Init)
	}
	if info.Processes != present(2) {
		t.Errorf("processes = %+v, want the two numeric directories of the fixture", info.Processes)
	}
	empty := Collect(context.Background(), Source{}, []Fact{FactInit, FactProcesses}, Options{})
	if empty.Init.State != Unknown || empty.Processes.State != Unknown {
		t.Errorf("without proc: %+v %+v", empty.Init, empty.Processes)
	}
}

func TestUsers(t *testing.T) {
	var utmp []byte
	utmp = append(utmp, utmpRecord(2, "reboot")...) // BOOT_TIME
	utmp = append(utmp, utmpRecord(7, "alice")...)
	utmp = append(utmp, utmpRecord(7, "alice")...)
	utmp = append(utmp, utmpRecord(7, "bob")...)
	utmp = append(utmp, utmpRecord(8, "carol")...) // DEAD_PROCESS
	utmp = append(utmp, 1, 2, 3)                   // truncated record
	tests := []struct {
		name string
		src  Source
		want Value[int]
	}{
		{"utmp", Source{FS: fstest.MapFS{"run/utmp": {Data: utmp}}}, present(2)},
		{"who", machine{commands: map[string]string{"who": "alice tty1 2024-01-01 10:00\nalice pts/0 2024-01-01 10:05\ndave pts/1 2024-01-01 11:00\n"}}.source(), present(2)},
		{"nothing", machine{}.source(), Value[int]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Collect(context.Background(), tt.src, []Fact{FactUsers}, Options{}).Users; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strings"
)

// FactBluetooth is whether Bluetooth is on.
const FactBluetooth Fact = "bluetooth"

func init() {
	collectors[FactBluetooth] = fact(collectBluetooth, func(i *Info) *Value[bool] { return &i.Bluetooth })
}

// collectBluetooth reports whether a Bluetooth adapter is registered and
// not blocked by rfkill. A blocked adapter may disappear from the
// Bluetooth class, so rfkill alone also proves that one exists. The fact
// is Absent without an adapter.
func collectBluetooth(_ context.Context, s *session) Value[bool] {
	adapters, _ := fs.ReadDir(s.fsys(), "sys/class/bluetooth")
	registered := false
	for _, e := range adapters {
		if strings.HasPrefix(e.Name(), "hci") {
			registered = true
		}
	}
	switches, _ := fs.ReadDir(s.fsys(), "sys/class/rfkill")
	hasSwitch, blocked := false, false
	for _, e := range switches {
		dir := path.Join("sys/class/rfkill", e.Name())
		if kind, _ := readTrimmed(s.fsys(), path.Join(dir, "type")); kind != "bluetooth" {
			continue
		}
		hasSwitch = true
		soft, _ := readTrimmed(s.fsys(), path.Join(dir, "soft"))
		hard, _ := readTrimmed(s.fsys(), path.Join(dir, "hard"))
		if soft == "1" || hard == "1" {
			blocked = true
		}
	}
	if !registered && !hasSwitch {
		return absent[bool]()
	}
	return present(registered && !blocked)
}

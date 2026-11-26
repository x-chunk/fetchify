package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strconv"
)

// FactBattery is the charge of every system battery.
const FactBattery Fact = "battery"

// Battery is the state of one system battery.
type Battery struct {
	// Percent is the charge in the range 0-100.
	Percent int
	// Status is the charging state as the kernel reports it: "Charging",
	// "Discharging", "Full", "Not charging" or "Unknown".
	Status string
}

func init() {
	collectors[FactBattery] = fact(collectBattery, func(i *Info) *Value[[]Battery] { return &i.Battery })
}

// collectBattery reads the power supplies of type Battery. Supplies with
// scope Device are batteries of peripherals such as mice and are left
// out. The fact is Absent without a battery and Unknown when a battery's
// charge cannot be read.
func collectBattery(_ context.Context, s *session) Value[[]Battery] {
	entries, err := fs.ReadDir(s.fsys(), "sys/class/power_supply")
	if err != nil {
		return absent[[]Battery]()
	}
	var batteries []Battery
	for _, e := range entries {
		dir := path.Join("sys/class/power_supply", e.Name())
		if kind, _ := readTrimmed(s.fsys(), path.Join(dir, "type")); kind != "Battery" {
			continue
		}
		if scope, _ := readTrimmed(s.fsys(), path.Join(dir, "scope")); scope == "Device" {
			continue
		}
		percent, ok := batteryPercent(s.fsys(), dir)
		if !ok {
			return Value[[]Battery]{}
		}
		status, err := readTrimmed(s.fsys(), path.Join(dir, "status"))
		if err != nil || status == "" {
			status = "Unknown"
		}
		batteries = append(batteries, Battery{Percent: percent, Status: status})
	}
	if len(batteries) == 0 {
		return absent[[]Battery]()
	}
	return present(batteries)
}

// batteryPercent reads the charge from capacity, or computes it from the
// energy or charge counters for drivers that do not provide capacity.
// The result is clamped to 0-100.
func batteryPercent(fsys fs.FS, dir string) (int, bool) {
	if n, ok := readInt(fsys, path.Join(dir, "capacity")); ok {
		return min(max(n, 0), 100), true
	}
	for _, prefix := range []string{"energy", "charge"} {
		now, ok1 := readInt(fsys, path.Join(dir, prefix+"_now"))
		full, ok2 := readInt(fsys, path.Join(dir, prefix+"_full"))
		if ok1 && ok2 && full > 0 {
			return min(max(int(int64(now)*100/int64(full)), 0), 100), true
		}
	}
	return 0, false
}

// readInt reads a file that holds one decimal integer.
func readInt(fsys fs.FS, name string) (int, bool) {
	s, err := readTrimmed(fsys, name)
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

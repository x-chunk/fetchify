package sysinfo

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
	"time"
)

// FactUptime is the time since boot.
const FactUptime Fact = "uptime"

func init() {
	collectors[FactUptime] = fact(collectUptime, func(i *Info) *Value[time.Duration] { return &i.Uptime })
}

// collectUptime reads the time since boot from proc/uptime.
func collectUptime(_ context.Context, s *session) Value[time.Duration] {
	data, err := fs.ReadFile(s.fsys(), "proc/uptime")
	if err != nil {
		return Value[time.Duration]{}
	}
	d, ok := parseUptime(string(data))
	if !ok {
		return Value[time.Duration]{}
	}
	return present(d)
}

// parseUptime parses the first field of proc/uptime, the seconds since
// boot with a fractional part, and truncates it to whole seconds.
func parseUptime(data string) (time.Duration, bool) {
	fields := strings.Fields(data)
	if len(fields) == 0 {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || seconds < 0 || seconds > float64(1<<40) {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

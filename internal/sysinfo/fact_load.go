package sysinfo

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
	"time"
)

// Facts about the processor load.
const (
	// FactCPUUsage is the share of processor time in use.
	FactCPUUsage Fact = "cpu_usage"
	// FactLoad is the load average.
	FactLoad Fact = "load"
)

// cpuUsageInterval is the time between the two samples of cpu_usage. It
// is a variable so that tests can shorten it; a run uses the value it had
// when the run started.
var cpuUsageInterval = 100 * time.Millisecond

func init() {
	collectors[FactCPUUsage] = fact(collectCPUUsage, func(i *Info) *Value[float64] { return &i.CPUUsage })
	collectors[FactLoad] = fact(collectLoad, func(i *Info) *Value[[3]float64] { return &i.Load })
}

// collectCPUUsage samples proc/stat twice and returns the share of time
// that was not idle in between, in percent. It is the only collector that
// sleeps, and it stops early when ctx is done.
func collectCPUUsage(ctx context.Context, s *session) Value[float64] {
	first, ok := readCPUTimes(s.fsys())
	if !ok {
		return Value[float64]{}
	}
	select {
	case <-ctx.Done():
		return Value[float64]{}
	case <-time.After(s.cpuUsageInterval):
	}
	second, ok := readCPUTimes(s.fsys())
	if !ok {
		return Value[float64]{}
	}
	percent, ok := cpuBusy(first, second)
	if !ok {
		return Value[float64]{}
	}
	return present(percent)
}

// cpuTimes holds the busy and the idle time of all processors, in clock
// ticks since boot.
type cpuTimes struct {
	busy, idle uint64
}

// readCPUTimes reads the aggregate "cpu" line of proc/stat.
func readCPUTimes(fsys fs.FS) (cpuTimes, bool) {
	data, err := fs.ReadFile(fsys, "proc/stat")
	if err != nil {
		return cpuTimes{}, false
	}
	return parseCPUTimes(string(data))
}

// parseCPUTimes parses the line "cpu user nice system idle iowait irq
// softirq steal guest guest_nice". Waiting for I/O counts as idle. The
// guest times are already part of user and nice and are not added again.
func parseCPUTimes(data string) (cpuTimes, bool) {
	line, _, _ := strings.Cut(data, "\n")
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuTimes{}, false
	}
	var t cpuTimes
	for i, f := range fields[1:min(len(fields), 9)] {
		n, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return cpuTimes{}, false
		}
		if i == 3 || i == 4 { // idle, iowait
			t.idle += n
		} else {
			t.busy += n
		}
	}
	return t, true
}

// cpuBusy returns the busy share between two samples in percent. It
// reports false when no time passed or the counters went backwards.
func cpuBusy(a, b cpuTimes) (float64, bool) {
	if b.busy < a.busy || b.idle < a.idle {
		return 0, false
	}
	busy, idle := b.busy-a.busy, b.idle-a.idle
	if busy+idle == 0 {
		return 0, false
	}
	return float64(busy) / float64(busy+idle) * 100, true
}

// collectLoad reads the 1, 5 and 15 minute load averages.
func collectLoad(_ context.Context, s *session) Value[[3]float64] {
	data, err := fs.ReadFile(s.fsys(), "proc/loadavg")
	if err != nil {
		return Value[[3]float64]{}
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return Value[[3]float64]{}
	}
	var load [3]float64
	for i := range load {
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return Value[[3]float64]{}
		}
		load[i] = v
	}
	return present(load)
}

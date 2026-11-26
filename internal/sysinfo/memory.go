package sysinfo

import "github.com/shirou/gopsutil/v3/mem"

// swap returns the swap usage, or nil when it cannot be read.
func swap() *Usage {
	stat, err := mem.SwapMemory()
	if err != nil {
		return nil
	}
	return &Usage{Used: stat.Used, Total: stat.Total, Available: stat.Free}
}

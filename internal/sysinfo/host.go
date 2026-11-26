package sysinfo

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
)

// processor returns the model of the first CPU and the logical CPU count.
func processor() CPU {
	result := CPU{Cores: runtime.NumCPU()}
	if stats, err := cpu.Info(); err == nil && len(stats) > 0 {
		result.Model = stats[0].ModelName
	}
	return result
}

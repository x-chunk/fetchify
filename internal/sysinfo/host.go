package sysinfo

import (
	"path"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
)

// shell returns the base name of the login shell named by the SHELL
// variable, or "" when the variable is not set.
func shell(getenv func(string) string) string {
	value := getenv("SHELL")
	if value == "" {
		return ""
	}
	return path.Base(value)
}

// processor returns the model of the first CPU and the logical CPU count.
func processor() CPU {
	result := CPU{Cores: runtime.NumCPU()}
	if stats, err := cpu.Info(); err == nil && len(stats) > 0 {
		result.Model = stats[0].ModelName
	}
	return result
}

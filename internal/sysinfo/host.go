package sysinfo

import (
	"os"
	"os/user"
	"path"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

// collectHost fills the identity fields of info: user, host name,
// operating system, kernel and uptime.
func collectHost(info *Info) {
	if current, err := user.Current(); err == nil {
		info.User = current.Username
	}
	if name, err := os.Hostname(); err == nil {
		info.Hostname = name
	}
	if stat, err := host.Info(); err == nil {
		info.OS = OS{Name: stat.Platform, Version: stat.PlatformVersion}
		info.Kernel = stat.KernelVersion
		info.Uptime = time.Duration(stat.Uptime) * time.Second
	}
}

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

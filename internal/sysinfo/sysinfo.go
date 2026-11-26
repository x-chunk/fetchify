package sysinfo

import (
	"os"
	"os/exec"
	"time"
)

// Info is a snapshot of the machine. Every field is optional: a zero
// value means the fact could not be determined.
type Info struct {
	// User is the identity the process runs as.
	User Value[User]
	// Hostname is the host name reported by the kernel.
	Hostname Value[string]
	// OS is the raw content of os-release. Parse it with
	// distro.ParseRelease.
	OS Value[[]byte]
	// Kernel is the kernel release, for example "6.8.0-45-generic".
	Kernel Value[string]
	// Packages lists the installed package counts per manager.
	Packages Value[[]PackageCount]
	// Uptime is the time since boot, with one-second resolution.
	Uptime Value[time.Duration]
	// Shell is the base name of the user's login shell.
	Shell Value[string]
	// Terminal is the name of the terminal emulator.
	Terminal Value[string]
	// Session is the protocol of the graphical session.
	Session Value[SessionType]
	// DE is the name of the desktop environment.
	DE Value[string]
	// WM is the name of the window manager or Wayland compositor.
	WM Value[string]
	// Host is the model of the machine.
	Host Value[string]
	// CPU describes the processor.
	CPU Value[CPU]
	// GPUs lists the names of the display adapters, in bus order.
	GPUs []string
	// Memory is the physical memory state, or nil when it cannot be read.
	Memory *Memory
	// Swap is the swap usage, or nil when it cannot be read. A Total of 0
	// means swap is disabled.
	Swap *Usage
	// Disks lists the mounted file systems worth showing, in mount order.
	Disks []Disk
}

// Usage is an amount of a resource in use out of a total.
type Usage struct {
	// Used and Total are sizes in bytes.
	Used, Total uint64
	// Percent is the used share in the range 0-100 as reported by the
	// system. It can differ from Used/Total for file systems, where
	// reserved blocks are excluded.
	Percent float64
}

// Memory is the state of physical memory.
type Memory struct {
	Usage
	// Types lists the distinct module types, for example "DDR4". It is
	// empty when the types cannot be read, which requires root.
	Types []string
}

// Disk is a mounted file system.
type Disk struct {
	Usage
	// Mountpoint is the path the file system is mounted at.
	Mountpoint string
	// Filesystem is the file system type, for example "ext4".
	Filesystem string
}

// runner runs an external command and returns its standard output.
// Collectors take it as a parameter so tests can replace it.
type runner func(name string, args ...string) ([]byte, error)

// run is the runner that executes real commands.
func run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// collectLegacy fills the fields that are not yet collected as facts.
// It may run external commands (lspci, xprop, wmctrl, dmidecode) and
// tolerates their absence.
func collectLegacy(info *Info) {
	info.GPUs = gpus(run)
	info.Memory = memory(run, os.Geteuid() == 0)
	info.Swap = swap()
	info.Disks = disks()
}

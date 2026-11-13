package sysinfo

import (
	"os"
	"os/exec"
	"time"
)

// Info is a snapshot of the machine. Every field is optional: a zero
// value means the fact could not be determined.
type Info struct {
	// User is the login name of the current user.
	User string
	// Hostname is the host name reported by the kernel.
	Hostname string
	// Root reports whether the process runs with superuser privileges.
	Root bool
	// OS identifies the operating system distribution.
	OS OS
	// Kernel is the kernel release, for example "6.8.0-45-generic".
	Kernel string
	// Uptime is the time since boot, with one-second resolution.
	Uptime time.Duration
	// Shell is the base name of the user's login shell.
	Shell string
	// Desktop describes the graphical session.
	Desktop Desktop
	// CPU describes the processor.
	CPU CPU
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

// OS identifies an operating system distribution.
type OS struct {
	// Name is the distribution identifier, for example "ubuntu".
	Name string
	// Version is the distribution version, for example "24.04".
	Version string
}

// Desktop describes the graphical session.
type Desktop struct {
	// WM is the name of the window manager or desktop environment. It is
	// empty when there is no graphical session or it cannot be detected.
	WM string
	// Server is the display server protocol of the session.
	Server DisplayServer
}

// DisplayServer is a display server protocol.
type DisplayServer string

// The display server protocols fetchify can tell apart.
const (
	X11     DisplayServer = "X11"
	Wayland DisplayServer = "Wayland"
)

// CPU describes the processor.
type CPU struct {
	// Model is the model name of the first processor.
	Model string
	// Cores is the number of logical CPUs usable by the process.
	Cores int
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

// Collect gathers a snapshot of the machine. It never fails: facts that
// cannot be determined are left at their zero values. It may run external
// commands (lspci, xprop, wmctrl, dmidecode) and tolerates their absence.
func Collect() Info {
	var info Info
	collectHost(&info)
	info.Root = os.Geteuid() == 0
	info.Shell = shell(os.Getenv)
	info.Desktop = desktop(run, os.Getenv)
	info.CPU = processor()
	info.GPUs = gpus(run)
	info.Memory = memory(run, info.Root)
	info.Swap = swap()
	info.Disks = disks()
	return info
}

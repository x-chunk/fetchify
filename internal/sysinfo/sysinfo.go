package sysinfo

import (
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
	// GPU lists the names of the display adapters in bus order, as
	// "<vendor> <device>" without cleaning.
	GPU Value[[]string]
	// Display lists the connected outputs.
	Display Value[[]Display]
	// Battery lists the system batteries.
	Battery Value[[]Battery]
	// Memory is the use of physical memory.
	Memory Value[Usage]
	// MemoryType lists the distinct types of the installed memory
	// modules, for example "DDR4".
	MemoryType Value[[]string]
	// Swap is the use of swap space.
	Swap Value[Usage]
	// Disks lists the mounted file systems worth showing, in mount order.
	Disks Value[[]Disk]
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
}

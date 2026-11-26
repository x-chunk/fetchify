package sysinfo

import (
	"net/netip"
	"time"
)

// Info is a snapshot of the machine: one field per fact, each with the
// state of its collection. Fields of facts that were not requested are
// Unknown. Values are raw data, such as sizes in bytes and unformatted
// names; presenting them is the job of the caller.
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

	// The facts below are opt-in: they are collected only when asked for.

	// LocalIP is the first global unicast address of the machine.
	LocalIP Value[LocalIP]
	// PublicIP is the address the machine has on the internet.
	PublicIP Value[netip.Addr]
	// Network is the interface of the default route.
	Network Value[Network]
	// Bluetooth reports whether Bluetooth is on.
	Bluetooth Value[bool]
	// CPUTemp is the temperature of the processor in degrees Celsius.
	CPUTemp Value[float64]
	// GPUTemp lists the temperatures of the GPUs that report one, in bus
	// order, in degrees Celsius.
	GPUTemp Value[[]float64]
	// CPUUsage is the share of processor time in use, in percent.
	CPUUsage Value[float64]
	// Load is the load average over 1, 5 and 15 minutes.
	Load Value[[3]float64]
	// GPUDriver lists the kernel drivers of the GPUs in bus order.
	GPUDriver Value[[]string]
	// GPUMemory lists the video memory use of the GPUs that report it,
	// in bus order.
	GPUMemory Value[[]Usage]
	// Board is the vendor and the name of the main board.
	Board Value[string]
	// BIOS identifies the firmware.
	BIOS Value[BIOS]
	// Init is the name of the init process.
	Init Value[string]
	// Processes is the number of processes.
	Processes Value[int]
	// Users is the number of distinct users with a login session.
	Users Value[int]
	// Locale is the locale of messages, for example "en_US.UTF-8".
	Locale Value[string]
	// Timezone is the IANA name of the time zone.
	Timezone Value[string]
	// Date is the current time.
	Date Value[time.Time]
	// Theme is the GTK or widget theme.
	Theme Value[string]
	// IconTheme is the icon theme.
	IconTheme Value[string]
	// Font is the interface font with its size.
	Font Value[string]
	// Cursor is the cursor theme.
	Cursor Value[string]
	// TerminalFont is the font of the terminal emulator.
	TerminalFont Value[string]
}

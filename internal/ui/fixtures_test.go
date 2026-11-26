package ui

import (
	"net/netip"
	"time"

	"go.xchunk.org/fetchify/internal/sysinfo"
)

// Fixture machines of the screen tests. They mirror the fixture trees of
// sysinfo.

const gibibyte = 1 << 30

// ubuntuDesktop is the machine of the target look of the plan.
func ubuntuDesktop() sysinfo.Info {
	return sysinfo.Info{
		User:     present(sysinfo.User{Name: "alice", UID: 1000}),
		Hostname: present("box"),
		OS:       present([]byte("ID=ubuntu\nID_LIKE=debian\nPRETTY_NAME=\"Ubuntu 24.04 LTS\"\n")),
		Kernel:   present("6.8.0-45-generic"),
		Uptime:   present(3*time.Hour + 25*time.Minute),
		Packages: present([]sysinfo.PackageCount{{Manager: "dpkg", Count: 1842}, {Manager: "flatpak", Count: 12}}),
		Shell:    present("zsh"),
		Terminal: present("kitty"),
		Session:  present(sysinfo.Wayland),
		DE:       present("GNOME"),
		WM:       present("Mutter"),
		Host:     present("ThinkPad X1 Carbon Gen 11"),
		CPU:      present(sysinfo.CPU{Model: "AMD Ryzen 7 5800X 8-Core Processor", Count: 16}),
		GPU:      present([]string{"NVIDIA Corporation GA104 [GeForce RTX 3070]"}),
		Display:  present([]sysinfo.Display{{Width: 2560, Height: 1440}}),
		Battery:  present([]sysinfo.Battery{{Percent: 78, Status: "Charging"}}),
		Memory:   present(sysinfo.Usage{Used: 8 * gibibyte, Total: 32 * gibibyte, Available: 24 * gibibyte}),
		Swap:     present(sysinfo.Usage{Used: 1 * gibibyte, Total: 4 * gibibyte, Available: 3 * gibibyte}),
		Disks: present([]sysinfo.Disk{
			{Mountpoint: "/", Filesystem: "ext4", Usage: sysinfo.Usage{Used: 100 * gibibyte, Total: 500 * gibibyte, Available: 376 * gibibyte}},
			{Mountpoint: "/home", Filesystem: "btrfs", Usage: sysinfo.Usage{Used: 250 * gibibyte, Total: 1000 * gibibyte, Available: 750 * gibibyte}},
		}),
	}
}

// archLaptop has a battery that runs low and two GPUs.
func archLaptop() sysinfo.Info {
	return sysinfo.Info{
		User:     present(sysinfo.User{Name: "bob", UID: 1000}),
		Hostname: present("archbook"),
		OS:       present([]byte("ID=arch\nNAME=\"Arch Linux\"\nPRETTY_NAME=\"Arch Linux\"\nBUILD_ID=rolling\n")),
		Kernel:   present("6.10.5-arch1-1"),
		Uptime:   present(26*time.Hour + 3*time.Minute),
		Packages: present([]sysinfo.PackageCount{{Manager: "pacman", Count: 1205}}),
		Shell:    present("bash"),
		Terminal: present("st"),
		Session:  present(sysinfo.X11),
		DE:       absent[string](),
		WM:       present("i3"),
		Host:     present("XPS 15 9520"),
		CPU:      present(sysinfo.CPU{Model: "12th Gen Intel(R) Core(TM) i7-12700H", Count: 20}),
		GPU: present([]string{
			"Intel Corporation Alder Lake-P GT2 [Iris Xe Graphics]",
			"NVIDIA Corporation GA107M [GeForce RTX 3050 Mobile]",
		}),
		Display: present([]sysinfo.Display{{Width: 1920, Height: 1200}, {Width: 2560, Height: 1440}}),
		Battery: present([]sysinfo.Battery{{Percent: 12, Status: "Discharging"}}),
		Memory:  present(sysinfo.Usage{Used: 14 * gibibyte, Total: 16 * gibibyte, Available: 2 * gibibyte}),
		Swap:    absent[sysinfo.Usage](),
		Disks: present([]sysinfo.Disk{
			{Mountpoint: "/", Filesystem: "ext4", Usage: sysinfo.Usage{Used: 450 * gibibyte, Total: 512 * gibibyte, Available: 40 * gibibyte}},
		}),
	}
}

// sshServer is a headless server, logged in as root, with an unknown
// terminal and a mount point of 1.8 TiB.
func sshServer() sysinfo.Info {
	return sysinfo.Info{
		User:     present(sysinfo.User{Name: "root", UID: 0}),
		Hostname: present("srv01"),
		OS:       present([]byte("ID=debian\nPRETTY_NAME=\"Debian GNU/Linux 12 (bookworm)\"\n")),
		Kernel:   present("6.1.0-25-amd64"),
		Uptime:   present(1111*time.Hour + 6*time.Minute),
		Packages: present([]sysinfo.PackageCount{{Manager: "dpkg", Count: 512}}),
		Shell:    present("bash"),
		Terminal: absent[string](),
		Session:  absent[sysinfo.SessionType](),
		DE:       absent[string](),
		WM:       absent[string](),
		Host:     present("PowerEdge R640"),
		CPU:      present(sysinfo.CPU{Model: "Intel(R) Xeon(R) Gold 6130 CPU @ 2.10GHz", Count: 64}),
		GPU:      present([]string{"102b 0536"}),
		Display:  absent[[]sysinfo.Display](),
		Battery:  absent[[]sysinfo.Battery](),
		Memory:   present(sysinfo.Usage{Used: 16 * gibibyte, Total: 64 * gibibyte, Available: 48 * gibibyte}),
		Swap:     present(sysinfo.Usage{Used: 0, Total: 2 * gibibyte, Available: 2 * gibibyte}),
		Disks: present([]sysinfo.Disk{
			{Mountpoint: "/", Filesystem: "xfs", Usage: sysinfo.Usage{Used: 20 * gibibyte, Total: 100 * gibibyte, Available: 80 * gibibyte}},
			{Mountpoint: "/srv", Filesystem: "xfs", Usage: sysinfo.Usage{Used: 1700 * gibibyte, Total: 1843 * gibibyte, Available: 143 * gibibyte}},
		}),
	}
}

// container has no DMI, no DRM, no os-release and an unknown user name.
func container() sysinfo.Info {
	return sysinfo.Info{
		User:     present(sysinfo.User{Name: "100123", UID: 100123}),
		Hostname: present("3f2a9c1b7d4e"),
		Kernel:   present("6.8.0-45-generic"),
		Uptime:   present(42 * time.Second),
		Packages: present([]sysinfo.PackageCount{{Manager: "apk", Count: 3}}),
		Terminal: absent[string](),
		Session:  absent[sysinfo.SessionType](),
		DE:       absent[string](),
		WM:       absent[string](),
		CPU:      present(sysinfo.CPU{Model: "AMD Ryzen 7 5800X 8-Core Processor", Count: 2}),
		GPU:      absent[[]string](),
		Display:  absent[[]sysinfo.Display](),
		Battery:  absent[[]sysinfo.Battery](),
		Memory:   present(sysinfo.Usage{Used: 512 << 20, Total: 900 << 20, Available: 388 << 20}),
		Swap:     absent[sysinfo.Usage](),
		Disks:    absent[[]sysinfo.Disk](),
	}
}

// raspberryPi is a board with a device tree and a Linux console.
func raspberryPi() sysinfo.Info {
	return sysinfo.Info{
		User:     present(sysinfo.User{Name: "pi", UID: 1000}),
		Hostname: present("raspberrypi"),
		OS:       present([]byte("ID=raspbian\nID_LIKE=debian\nPRETTY_NAME=\"Raspbian GNU/Linux 11 (bullseye)\"\n")),
		Kernel:   present("6.6.31+rpt-rpi-v8"),
		Uptime:   present(10 * time.Minute),
		Packages: present([]sysinfo.PackageCount{{Manager: "dpkg", Count: 3}}),
		Shell:    present("bash"),
		Terminal: absent[string](),
		Session:  absent[sysinfo.SessionType](),
		DE:       absent[string](),
		WM:       absent[string](),
		Host:     present("Raspberry Pi 4 Model B Rev 1.4"),
		CPU:      present(sysinfo.CPU{Model: "BCM2835", Count: 4}),
		GPU:      absent[[]string](),
		Display:  absent[[]sysinfo.Display](),
		Battery:  absent[[]sysinfo.Battery](),
		Memory:   present(sysinfo.Usage{Used: 900 << 20, Total: 3793 << 20, Available: 2893 << 20}),
		Swap:     present(sysinfo.Usage{Used: 0, Total: 100 << 20, Available: 100 << 20}),
		Disks: present([]sysinfo.Disk{
			{Mountpoint: "/", Filesystem: "ext4", Usage: sysinfo.Usage{Used: 6 * gibibyte, Total: 29 * gibibyte, Available: 22 * gibibyte}},
		}),
	}
}

// everyOptIn has a value for every opt-in fact.
func everyOptIn() sysinfo.Info {
	info := ubuntuDesktop()
	info.LocalIP = present(sysinfo.LocalIP{Addr: netip.MustParseAddr("192.168.1.23"), Interface: "wlan0"})
	info.PublicIP = present(netip.MustParseAddr("203.0.113.7"))
	info.Network = present(sysinfo.Network{Interface: "wlan0", Type: "wifi", SSID: "Home"})
	info.Bluetooth = present(true)
	info.CPUTemp = present(54.4)
	info.GPUTemp = present([]float64{71, 90})
	info.CPUUsage = present(31.2)
	info.Load = present([3]float64{0.52, 0.58, 0.59})
	info.GPUDriver = present([]string{"nvidia"})
	info.GPUMemory = present([]sysinfo.Usage{{Used: 512 << 20, Total: 8 * gibibyte, Available: 8*gibibyte - 512<<20}})
	info.Board = present("ASUSTeK COMPUTER INC. P8H61-M LX")
	info.BIOS = present(sysinfo.BIOS{Vendor: "American Megatrends Inc.", Version: "4601", Date: "09/18/2013"})
	info.Init = present("systemd")
	info.Processes = present(308)
	info.Users = present(1)
	info.Locale = present("en_US.UTF-8")
	info.Timezone = present("Asia/Oral")
	info.Date = present(time.Date(2026, 11, 26, 16, 6, 0, 0, time.UTC))
	info.Theme = present("Adwaita-dark")
	info.IconTheme = present("Papirus")
	info.Font = present("Inter 10")
	info.Cursor = present("Bibata")
	info.TerminalFont = present("JetBrains Mono")
	info.Media = present(sysinfo.Media{Artist: "Daft Punk", Title: "Around the World", Status: "Playing"})
	return info
}

// unknownEverything has every default fact Unknown.
func unknownEverything() sysinfo.Info {
	return sysinfo.Info{}
}

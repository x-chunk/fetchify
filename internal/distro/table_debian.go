package distro

import "go.xchunk.org/beautify/dra"

// The Debian family, Ubuntu and its derivatives.
func init() {
	register(
		Distro{
			ID:     "debian",
			Name:   "Debian GNU/Linux",
			Colors: []dra.Color{hex("#d70a53"), dra.TerminalDefault},
			Accent: hex("#d70a53"),
			Icon:   '', // nf-linux-debian
		},
		Distro{
			ID:     "ubuntu",
			Name:   "Ubuntu",
			Colors: []dra.Color{hex("#e95420"), dra.TerminalDefault},
			Accent: hex("#e95420"),
			Icon:   '', // nf-linux-ubuntu
		},
		Distro{
			ID:     "linuxmint",
			Name:   "Linux Mint",
			Colors: []dra.Color{hex("#87cf3e"), dra.TerminalDefault},
			Accent: hex("#87cf3e"),
			Icon:   '', // nf-linux-linuxmint
		},
		Distro{
			ID:     "pop",
			Name:   "Pop!_OS",
			Colors: []dra.Color{hex("#48b9c7"), dra.TerminalDefault},
			Accent: hex("#48b9c7"),
			Icon:   '', // nf-linux-pop_os
		},
		Distro{
			ID:     "elementary",
			Name:   "elementary OS",
			Colors: []dra.Color{dra.TerminalDefault},
			Accent: hex("#64baff"),
			Icon:   '', // nf-linux-elementary
		},
		Distro{
			ID:     "zorin",
			Name:   "Zorin OS",
			Colors: []dra.Color{hex("#15a6f0")},
			Accent: hex("#15a6f0"),
			Icon:   '', // nf-linux-zorin
		},
		Distro{
			ID:     "kali",
			Name:   "Kali Linux",
			Colors: []dra.Color{hex("#367bf0"), dra.TerminalDefault},
			Accent: hex("#367bf0"),
			Icon:   '', // nf-linux-kali_linux
		},
	)
	alias("debian", "raspbian")
}

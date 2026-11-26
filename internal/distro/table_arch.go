package distro

import "go.xchunk.org/beautify/dra"

// The Arch Linux family.
func init() {
	register(
		Distro{
			ID:     "arch",
			Name:   "Arch Linux",
			Colors: []dra.Color{hex("#1793d1")},
			Accent: hex("#1793d1"),
			Icon:   '', // nf-linux-archlinux
		},
		Distro{
			ID:     "manjaro",
			Name:   "Manjaro",
			Colors: []dra.Color{hex("#35bf5c")},
			Accent: hex("#35bf5c"),
			Icon:   '', // nf-linux-manjaro
		},
		Distro{
			ID:     "endeavouros",
			Name:   "EndeavourOS",
			Colors: []dra.Color{hex("#8f5bd6"), hex("#f25c54"), hex("#5b8def")},
			Accent: hex("#8f5bd6"),
			Icon:   '', // nf-linux-endeavour
		},
		Distro{
			ID:     "cachyos",
			Name:   "CachyOS",
			Colors: []dra.Color{hex("#00b39f"), hex("#6fe3d0")},
			Accent: hex("#00b39f"),
			Icon:   '', // nf-linux-cachyos
		},
		Distro{
			ID:     "artix",
			Name:   "Artix Linux",
			Colors: []dra.Color{hex("#10a0cc")},
			Accent: hex("#10a0cc"),
			Icon:   '', // nf-linux-artix
		},
	)
	alias("arch", "archarm")
}

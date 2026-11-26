package distro

import "go.xchunk.org/beautify/dra"

// Distributions that are not derived from another one in the table.
func init() {
	register(
		Distro{
			ID:     "opensuse",
			Name:   "openSUSE",
			Colors: []dra.Color{hex("#73ba25"), dra.TerminalDefault},
			Accent: hex("#73ba25"),
			Icon:   '', // nf-linux-opensuse
		},
		Distro{
			ID:     "nixos",
			Name:   "NixOS",
			Colors: []dra.Color{hex("#5277c3"), hex("#7ebae4")},
			Accent: hex("#7ebae4"),
			Icon:   '', // nf-linux-nixos
		},
		Distro{
			ID:     "gentoo",
			Name:   "Gentoo",
			Colors: []dra.Color{hex("#9e86c8"), dra.TerminalDefault},
			Accent: hex("#9e86c8"),
			Icon:   '', // nf-linux-gentoo
		},
		Distro{
			ID:     "void",
			Name:   "Void Linux",
			Colors: []dra.Color{hex("#62b57a"), hex("#7f7f7f")},
			Accent: hex("#62b57a"),
			Icon:   '', // nf-linux-void
		},
		Distro{
			ID:     "alpine",
			Name:   "Alpine Linux",
			Colors: []dra.Color{hex("#2f8fcf"), dra.TerminalDefault},
			Accent: hex("#2f8fcf"),
			Icon:   '', // nf-linux-alpine
		},
	)
	alias("opensuse", "opensuse-tumbleweed", "opensuse-leap", "opensuse-slowroll", "sles")
}

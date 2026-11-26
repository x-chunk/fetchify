package distro

import "go.xchunk.org/beautify/dra"

// The Red Hat family.
func init() {
	register(
		Distro{
			ID:     "fedora",
			Name:   "Fedora Linux",
			Colors: []dra.Color{hex("#51a2da"), dra.TerminalDefault},
			Accent: hex("#51a2da"),
			Icon:   '', // nf-linux-fedora
		},
		Distro{
			ID:     "rhel",
			Name:   "Red Hat Enterprise Linux",
			Colors: []dra.Color{hex("#ee0000")},
			Accent: hex("#ee0000"),
			Icon:   '', // nf-linux-redhat
		},
		Distro{
			ID:   "centos",
			Name: "CentOS",
			Colors: []dra.Color{
				hex("#efa724"), hex("#8eb929"), hex("#4f6fd8"), hex("#b54fa0"), dra.TerminalDefault,
			},
			Accent: hex("#b54fa0"),
			Icon:   '', // nf-linux-centos
		},
		Distro{
			ID:     "rocky",
			Name:   "Rocky Linux",
			Colors: []dra.Color{hex("#10b981")},
			Accent: hex("#10b981"),
			Icon:   '', // nf-linux-rocky_linux
		},
		Distro{
			ID:   "almalinux",
			Name: "AlmaLinux",
			Colors: []dra.Color{
				hex("#ff4649"), hex("#ffcb12"), hex("#0069da"), hex("#7cc33f"), hex("#24c2ff"),
			},
			Accent: hex("#24c2ff"),
			Icon:   '', // nf-linux-almalinux
		},
	)
	alias("rhel", "ol")
}

package ui

// Icons are Nerd Font glyphs from the stable ranges of Nerd Fonts v3,
// checked against its glyphnames.json. Every icon counts as one cell.

// itemIcons maps item identifiers to their icons. The os item takes the
// icon of the distribution instead.
var itemIcons = map[string]rune{
	"kernel":   '',          // nf-fa-linux
	"uptime":   '\U000f0150', // nf-md-clock_outline
	"packages": '\U000f03d6', // nf-md-package_variant
	"shell":    '',          // nf-oct-terminal
	"terminal": '',          // nf-fa-terminal
	"de":       '',          // nf-fa-desktop
	"wm":       '',          // nf-fa-window_maximize
	"host":     '\U000f0322', // nf-md-laptop
	"cpu":      '',          // nf-oct-cpu
	"gpu":      '\U000f08ae', // nf-md-expansion_card
	"display":  '\U000f0379', // nf-md-monitor
	"battery":  '\U000f0079', // nf-md-battery
	"memory":   '\U000f035b', // nf-md-memory
	"swap":     '\U000f04e1', // nf-md-swap_horizontal
	"disks":    '\U000f02ca', // nf-md-harddisk
}

// groupIcons maps the titles of the default groups to their icons. A
// group with icon = "auto" and another title has no icon.
var groupIcons = map[string]rune{
	"System":   '\U000f02fd', // nf-md-information_outline
	"Hardware": '\U000f061a', // nf-md-chip
	"Storage":  '\U000f01bc', // nf-md-database
}

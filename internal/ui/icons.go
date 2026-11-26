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

	"local_ip":      '\U000f0a60', // nf-md-ip_network
	"public_ip":     '\U000f01e7', // nf-md-earth
	"network":       '\U000f0002', // nf-md-access_point_network
	"bluetooth":     '\U000f00af', // nf-md-bluetooth
	"cpu_temp":      '\U000f050f', // nf-md-thermometer
	"gpu_temp":      '\U000f050f', // nf-md-thermometer
	"cpu_usage":     '\U000f029a', // nf-md-gauge
	"load":          '\U000f012a', // nf-md-chart_line
	"gpu_driver":    '\U000f08bb', // nf-md-cog_outline
	"gpu_memory":    '\U000f0fb2', // nf-md-expansion_card_variant
	"board":         '\U000f0697', // nf-md-developer_board
	"bios":          '\U000f062e', // nf-md-tune
	"init":          '\U000f14de', // nf-md-rocket_launch
	"processes":     '\U000f027b', // nf-md-format_list_numbered
	"users":         '\U000f000e', // nf-md-account_multiple
	"locale":        '\U000f05ca', // nf-md-translate
	"timezone":      '\U000f06cd', // nf-md-earth_box
	"date":          '\U000f00ed', // nf-md-calendar
	"theme":         '\U000f03d8', // nf-md-palette
	"icon_theme":    '\U000f069d', // nf-md-folder_star
	"font":          '\U000f06d6', // nf-md-format_font
	"cursor":        '\U000f01c0', // nf-md-cursor_default
	"terminal_font": '\U000f06d6', // nf-md-format_font
	"media":         '\U000f075a', // nf-md-music
}

// groupIcons maps the titles of the default groups to their icons. A
// group with icon = "auto" and another title has no icon.
var groupIcons = map[string]rune{
	"System":   '\U000f02fd', // nf-md-information_outline
	"Hardware": '\U000f061a', // nf-md-chip
	"Storage":  '\U000f01bc', // nf-md-database
}

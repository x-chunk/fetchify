package config

import (
	"slices"
	"time"

	"go.xchunk.org/fetchify/v2/internal/sysinfo"
)

// Config is the complete configuration of a run. Values are already
// validated: enumerations hold one of their constants and colors are
// "#rgb" or "#rrggbb".
type Config struct {
	// Clear clears the screen before printing.
	Clear bool
	// Footer shows the line under the card.
	Footer bool
	// Timeout is the default time limit of every item.
	Timeout time.Duration
	Logo    Logo
	Colors  Colors
	Border  Border
	Icons   Icons
	Bars    Bars
	Palette Palette
	// Groups are the sections of the card in display order.
	Groups []Group
	// Items holds the settings of single items by identifier. Items
	// without settings have no entry.
	Items map[string]Item
}

// Logo configures the logo.
type Logo struct {
	// Name is LogoAuto, LogoNone or the identifier of a logo.
	Name string
	Size LogoSize
	// Gradient recolors the logo with a gradient through Colors, or
	// through the logo's own colors when Colors is empty.
	Gradient  bool
	Colors    []string
	Direction Direction
}

// Colors configures the color output.
type Colors struct {
	Mode ColorMode
	// Accent is AccentAuto or a color.
	Accent string
}

// Border configures the frame of the card.
type Border struct {
	Style BorderStyle
	Color BorderColor
	// Colors are the gradient stops of BorderCustom.
	Colors    []string
	Direction Direction
}

// Icons configures the Nerd Font icons.
type Icons struct {
	Enabled bool
}

// Bars configures the usage bars.
type Bars struct {
	Style BarStyle
	// Width is the number of cells of a bar.
	Width int
	// Thresholds are the percentages where the color of a bar turns to
	// warning and to danger.
	Thresholds [2]int
}

// Palette configures the color palette.
type Palette struct {
	Style PaletteStyle
}

// Group is a section of the card.
type Group struct {
	Title string
	// Icon is IconAuto, "" for none, or the icon itself.
	Icon string
	// Items lists item identifiers in display order.
	Items []string
}

// Item holds the settings of one item. Nil and zero fields keep the
// defaults of the item.
type Item struct {
	// Label replaces the label.
	Label *string
	// Icon replaces the icon; an empty string removes it.
	Icon *string
	// Timeout replaces the time limit of the item's facts.
	Timeout time.Duration
	// Ignore lists the mount points the disks item leaves out.
	Ignore []string
	// ShowType adds the module types to the memory item.
	ShowType bool
	// URL is the endpoint of the public_ip item.
	URL string
}

// Special values of string settings.
const (
	// LogoAuto chooses the logo of the running distribution.
	LogoAuto = "auto"
	// LogoNone shows no logo.
	LogoNone = "none"
	// AccentAuto takes the accent color of the distribution.
	AccentAuto = "auto"
	// IconAuto takes the icon registered for a default group of the same
	// title.
	IconAuto = "auto"
)

// LogoSize selects the size of the logo.
type LogoSize string

// Logo sizes.
const (
	SizeAuto  LogoSize = "auto"
	SizeFull  LogoSize = "full"
	SizeSmall LogoSize = "small"
)

// Direction is the axis of a gradient.
type Direction string

// Gradient directions.
const (
	Vertical   Direction = "vertical"
	Horizontal Direction = "horizontal"
	Diagonal   Direction = "diagonal"
)

// ColorMode selects the color profile.
type ColorMode string

// Color modes. ModeAuto detects the terminal and honors NO_COLOR.
const (
	ModeAuto      ColorMode = "auto"
	ModeTrueColor ColorMode = "truecolor"
	Mode256       ColorMode = "256"
	Mode16        ColorMode = "16"
	ModeNone      ColorMode = "none"
)

// BorderStyle selects the characters of the frame.
type BorderStyle string

// Border styles.
const (
	BorderRounded BorderStyle = "rounded"
	BorderSquare  BorderStyle = "square"
	BorderDouble  BorderStyle = "double"
	BorderHeavy   BorderStyle = "heavy"
	BorderASCII   BorderStyle = "ascii"
)

// BorderColor selects how the frame is colored.
type BorderColor string

// Border colors.
const (
	BorderAccent  BorderColor = "accent"
	BorderLogo    BorderColor = "logo"
	BorderRainbow BorderColor = "rainbow"
	BorderCustom  BorderColor = "custom"
)

// BarStyle selects the characters of the usage bars.
type BarStyle string

// Bar styles.
const (
	BarBlock BarStyle = "block"
	BarThin  BarStyle = "thin"
)

// PaletteStyle selects how the terminal palette is shown.
type PaletteStyle string

// Palette styles.
const (
	PaletteDots   PaletteStyle = "dots"
	PaletteBlocks PaletteStyle = "blocks"
	PaletteNone   PaletteStyle = "none"
)

// Defaults returns the configuration that applies without a file. Every
// call returns a new value that the caller may modify.
func Defaults() Config {
	return Config{
		Footer:  true,
		Timeout: sysinfo.DefaultTimeout,
		Logo:    Logo{Name: LogoAuto, Size: SizeAuto, Colors: []string{}, Direction: Vertical},
		Colors:  Colors{Mode: ModeAuto, Accent: AccentAuto},
		Border:  Border{Style: BorderRounded, Color: BorderAccent, Colors: []string{}, Direction: Vertical},
		Icons:   Icons{Enabled: true},
		Bars:    Bars{Style: BarBlock, Width: 16, Thresholds: [2]int{60, 85}},
		Palette: Palette{Style: PaletteDots},
		Groups: []Group{
			{Title: "System", Icon: IconAuto, Items: []string{"os", "kernel", "uptime", "packages", "shell", "terminal", "de", "wm"}},
			{Title: "Hardware", Icon: IconAuto, Items: []string{"host", "cpu", "gpu", "display", "battery"}},
			{Title: "Storage", Icon: IconAuto, Items: []string{"memory", "swap", "disks"}},
		},
		Items: map[string]Item{
			"disks": {Ignore: slices.Clone(sysinfo.DefaultDiskIgnore)},
		},
	}
}

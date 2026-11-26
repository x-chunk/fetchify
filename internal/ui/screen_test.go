package ui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/golden"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// renderScreen renders the screen of info and cfg at profile p. A width
// of 0 is unbounded.
func renderScreen(t *testing.T, info sysinfo.Info, cfg config.Config, p dra.Profile, width int) string {
	t.Helper()
	ctx := beautify.NewContext(nil)
	if width > 0 {
		ctx = ctx.WithWidth(width)
	}
	m, err := Screen(Input{Info: info, Config: cfg, Elapsed: 12345 * time.Microsecond}).Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return dra.NewEngine(p).String(m)
}

// screenCase is a golden test of the whole screen.
type screenCase struct {
	name string
	info sysinfo.Info
	cfg  config.Config
}

// screenCases are the golden tests of the screen. Every case is rendered
// without styles and in true color.
func screenCases() []screenCase {
	noIcons := config.Defaults()
	noIcons.Icons.Enabled = false
	blocksPalette := config.Defaults()
	blocksPalette.Palette.Style = config.PaletteBlocks
	noPalette := config.Defaults()
	noPalette.Palette.Style = config.PaletteNone
	koala := config.Defaults()
	koala.Logo.Name = "koala"
	small := config.Defaults()
	small.Logo.Size = config.SizeSmall
	noLogo := config.Defaults()
	noLogo.Logo.Name = config.LogoNone
	logoGradient := config.Defaults()
	logoGradient.Logo.Gradient = true
	logoGradient.Logo.Direction = config.Horizontal
	logoGradient.Border.Color = config.BorderLogo
	rainbowFrame := config.Defaults()
	rainbowFrame.Border.Color = config.BorderRainbow
	rainbowFrame.Border.Direction = config.Diagonal
	rainbowFrame.Logo.Gradient = true
	rainbowFrame.Logo.Colors = []string{"#ff0000", "#0000ff"}
	customFrame := config.Defaults()
	customFrame.Border.Color = config.BorderCustom
	customFrame.Border.Colors = []string{"#00ff00"}
	noFooter := config.Defaults()
	noFooter.Footer = false
	optIn := config.Defaults()
	var optInItems []string
	for _, id := range ItemIDs() {
		if !slices.ContainsFunc(optIn.Groups, func(g config.Group) bool { return slices.Contains(g.Items, id) }) {
			optInItems = append(optInItems, id)
		}
	}
	optIn.Groups = append(optIn.Groups, config.Group{Title: "Extra", Icon: "", Items: optInItems})
	custom := config.Defaults()
	custom.Groups = []config.Group{
		{Title: "Me", Icon: config.IconAuto, Items: []string{"os", "shell", "uptime"}},
		{Title: "Nothing", Items: []string{"battery"}},
		{Title: "Box", Icon: "*", Items: []string{"cpu", "memory"}},
	}
	thin := config.Defaults()
	thin.Bars.Style = config.BarThin
	thin.Border.Style = config.BorderHeavy
	return []screenCase{
		{"ubuntu", ubuntuDesktop(), config.Defaults()},
		{"arch", archLaptop(), config.Defaults()},
		{"ssh", sshServer(), config.Defaults()},
		{"container", container(), config.Defaults()},
		{"raspberry", raspberryPi(), config.Defaults()},
		{"unknown", unknownEverything(), config.Defaults()},
		{"no_icons", ubuntuDesktop(), noIcons},
		{"palette_blocks", ubuntuDesktop(), blocksPalette},
		{"palette_none", ubuntuDesktop(), noPalette},
		{"logo_koala", ubuntuDesktop(), koala},
		{"logo_small", archLaptop(), small},
		{"logo_none", ubuntuDesktop(), noLogo},
		{"gradient_logo", ubuntuDesktop(), logoGradient},
		{"gradient_rainbow", archLaptop(), rainbowFrame},
		{"frame_custom_solid", archLaptop(), customFrame},
		{"no_footer", ubuntuDesktop(), noFooter},
		{"opt_in", everyOptIn(), optIn},
		{"custom_groups", sshServer(), custom},
		{"thin_heavy", archLaptop(), thin},
	}
}

func TestScreenGoldenFiles(t *testing.T) {
	for _, c := range screenCases() {
		t.Run(c.name, func(t *testing.T) {
			plain := renderScreen(t, c.info, c.cfg, dra.ProfilePlain, 0)
			golden.Assert(t, "screen_"+c.name+"_plain", []byte(plain+"\n"))
			golden.Assert(t, "screen_"+c.name+"_truecolor", []byte(renderScreen(t, c.info, c.cfg, dra.ProfileTrueColor, 0)+"\n"))
			assertCardAligned(t, plain)
		})
	}
}

// assertCardAligned checks that every line of the card ends at the same
// column, which is where its right edge is.
func assertCardAligned(t *testing.T, plain string) {
	t.Helper()
	ends := map[int]int{}
	for _, line := range strings.Split(plain, "\n") {
		trimmed := strings.TrimRight(line, " ")
		if r := []rune(trimmed); len(r) > 0 && strings.ContainsRune("╮│┤╯┐┘╗║╣╝┓┃┫┛+|", r[len(r)-1]) {
			ends[len(r)]++
		}
	}
	if len(ends) > 1 {
		t.Errorf("card lines end at different columns: %v\n%s", ends, plain)
	}
}

func TestScreenWithoutGroups(t *testing.T) {
	cfg := config.Defaults()
	cfg.Groups = nil
	cfg.Logo.Name = config.LogoNone
	if got := renderScreen(t, ubuntuDesktop(), cfg, dra.ProfilePlain, 0); got != "" {
		t.Errorf("no groups and no logo: %q", got)
	}
	cfg.Logo.Name = "koala"
	if got := renderScreen(t, ubuntuDesktop(), cfg, dra.ProfilePlain, 0); !strings.Contains(got, `/"  "\`) || strings.Contains(got, "╭") {
		t.Errorf("no groups draws the logo only:\n%s", got)
	}
}

func TestCaption(t *testing.T) {
	user := present(sysinfo.User{Name: "a"})
	host := present("h")
	tests := []struct {
		info sysinfo.Info
		want string
	}{
		{sysinfo.Info{User: user, Hostname: host}, "a@h"},
		{sysinfo.Info{User: user}, "a"},
		{sysinfo.Info{Hostname: host}, "h"},
		{sysinfo.Info{}, ""},
	}
	for _, tt := range tests {
		if got := caption(tt.info); got != tt.want {
			t.Errorf("caption = %q, want %q", got, tt.want)
		}
	}
}

func TestItemSettings(t *testing.T) {
	cfg := config.Defaults()
	label, icon := "Processor", ""
	cfg.Items["cpu"] = config.Item{Label: &label, Icon: &icon}
	cfg.Groups = []config.Group{{Title: "Mine", Items: []string{"cpu", "battery"}}}
	got := renderScreen(t, ubuntuDesktop(), cfg, dra.ProfilePlain, 0)
	if !strings.Contains(got, "Processor  AMD Ryzen 7 5800X") || !strings.Contains(got, "Mine") {
		t.Errorf("settings not applied:\n%s", got)
	}
}

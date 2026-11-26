package ui

import (
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// Input is everything the screen shows. ui measures and collects
// nothing; the caller passes the results in.
type Input struct {
	// Info holds the collected facts.
	Info sysinfo.Info
	// Config is the validated configuration.
	Config config.Config
	// Elapsed is the time the collection took, shown in the footer.
	Elapsed time.Duration
}

// Screen returns the whole fetch screen for in.
func Screen(in Input) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		s := newScreen(in)
		return s.column(s.fullLayout()).Render(ctx)
	})
}

// screen holds what every layout of the screen shares.
type screen struct {
	in     Input
	distro distro.Distro
	theme  theme
	groups []group
}

// newScreen resolves the distribution, the theme and the rows of in.
func newScreen(in Input) screen {
	d := distro.Lookup(distro.ParseRelease(in.Info.OS.V))
	root := in.Info.User.State == sysinfo.Present && in.Info.User.V.UID == 0
	env := itemEnv{cfg: in.Config, distro: d}
	return screen{
		in:     in,
		distro: d,
		theme:  newTheme(in.Config, d, root),
		groups: buildGroups(in.Info, env),
	}
}

// fullLayout is the layout with every detail.
func (s screen) fullLayout() layout {
	cfg := s.in.Config
	return layout{
		icons:      cfg.Icons.Enabled,
		details:    true,
		bars:       true,
		barWidth:   cfg.Bars.Width,
		barStyle:   cfg.Bars.Style,
		thresholds: cfg.Bars.Thresholds,
	}
}

// card returns the card in the given layout.
func (s screen) card(lay layout) card {
	return card{
		groups:  s.groups,
		theme:   s.theme,
		layout:  lay,
		border:  s.in.Config.Border.Style,
		caption: caption(s.in.Info),
		dots:    s.in.Config.Palette.Style == config.PaletteDots,
	}
}

// column stacks the card and what hangs under it. Without groups there
// is no card, and nothing hangs under nothing.
func (s screen) column(lay layout) beautify.Composable {
	if len(s.groups) == 0 {
		return blocks.Space(0, 0)
	}
	children := []beautify.Composable{s.card(lay)}
	if s.in.Config.Palette.Style == config.PaletteBlocks {
		children = append(children, paletteBlocks())
	}
	return blocks.Column(children)
}

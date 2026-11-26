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

// Screen returns the whole fetch screen for in. It adapts to the width
// of the context: see screen.render.
func Screen(in Input) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		m, _, err := newScreen(in).render(ctx)
		return m, err
	})
}

// step names a layout of the screen, from the richest to the poorest.
type step uint8

// The layout steps of section 13.11 of the plan.
const (
	stepFullLogo step = iota + 1
	stepSmallLogo
	stepNoLogo
	stepNoDetails
	stepNarrowBars
	stepNoBars
	stepTruncated
)

// minBarWidth is the narrowest bar the screen shrinks bars to before it
// drops them.
const minBarWidth = 6

// candidate is a layout of the screen to try.
type candidate struct {
	step  step
	build beautify.Composable
}

// candidates returns the layouts to try in order: the logo sizes the
// configuration allows, then no logo, then a card with less and less
// detail.
func (s screen) candidates() []candidate {
	var list []candidate
	full := s.fullLayout()
	if _, ok := s.logoDistro(); ok {
		switch s.in.Config.Logo.Size {
		case config.SizeFull:
			list = append(list, candidate{stepFullLogo, s.withLogo(distro.Full, s.column(full))})
		case config.SizeSmall:
			list = append(list, candidate{stepSmallLogo, s.withLogo(distro.Small, s.column(full))})
		default:
			list = append(list,
				candidate{stepFullLogo, s.withLogo(distro.Full, s.column(full))},
				candidate{stepSmallLogo, s.withLogo(distro.Small, s.column(full))})
		}
	}
	list = append(list, candidate{stepNoLogo, s.column(full)})
	lay := full
	lay.details = false
	list = append(list, candidate{stepNoDetails, s.column(lay)})
	for w := lay.barWidth - 1; w >= minBarWidth; w-- {
		lay.barWidth = w
		list = append(list, candidate{stepNarrowBars, s.column(lay)})
	}
	lay.bars = false
	list = append(list, candidate{stepNoBars, s.column(lay)})
	lay.truncate = true
	list = append(list, candidate{stepTruncated, s.column(lay)})
	return list
}

// render draws the first candidate that fits the width of ctx, measured
// by rendering it without a bound. In an unbounded context that is the
// richest one. When nothing fits, the last candidate, whose values are
// truncated, is rendered within the bound, and beautify cuts what still
// sticks out.
func (s screen) render(ctx beautify.Context) (dra.Model, step, error) {
	list := s.candidates()
	bound, bounded := ctx.Width()
	if !bounded {
		m, err := list[0].build.Render(ctx)
		return m, list[0].step, err
	}
	for _, c := range list[:len(list)-1] {
		m, err := c.build.Render(ctx.Unbounded())
		if err != nil {
			return dra.Model{}, c.step, err
		}
		if m.Width() <= bound {
			return m, c.step, nil
		}
	}
	last := list[len(list)-1]
	m, err := last.build.Render(ctx)
	if err == nil && m.Width() > bound {
		m = dra.Truncate(m, bound, dra.Model{})
	}
	return m, last.step, err
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
	if s.in.Config.Footer {
		children = append(children, footer(s.theme, s.in.Elapsed))
	}
	return blocks.Column(children)
}

// withLogo places the logo of the given size left of column, both
// centered vertically against each other. Without a logo the column
// stands alone.
func (s screen) withLogo(size distro.Size, column beautify.Composable) beautify.Composable {
	d, ok := s.logoDistro()
	if !ok {
		return column
	}
	logo := beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		m := logoModel(ctx.Factory(), d, d.Logo(size))
		if cfg := s.in.Config.Logo; cfg.Gradient {
			stops := hexColors(cfg.Colors)
			if len(stops) == 0 {
				stops = solidColors(d.Colors)
			}
			m = recolorLogo(m, dra.NewGradient(axes[cfg.Direction], stops...))
		}
		return blocks.Model(m).Render(ctx)
	})
	return blocks.Row(
		[]blocks.Cell{blocks.Auto(logo), blocks.Auto(column)},
		blocks.RowGap(logoSpacing),
		blocks.RowAlign(dra.AlignMiddle),
	)
}

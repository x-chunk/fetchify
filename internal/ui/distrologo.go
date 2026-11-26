package ui

import (
	"strings"
	"unicode/utf8"

	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
)

// logoSpacing is the number of cells between the logo and the card.
const logoSpacing = 2

// logoDistro returns the distribution whose logo the configuration asks
// for and whether there is a logo at all.
func (s screen) logoDistro() (distro.Distro, bool) {
	switch name := s.in.Config.Logo.Name; name {
	case config.LogoNone:
		return distro.Distro{}, false
	case config.LogoAuto, "":
		return s.distro, true
	default:
		d, ok := distro.ByID(name)
		return d, ok
	}
}

// logoModel turns a logo into one model whose palette is the colors of
// the distribution and whose colormap follows the runs of the logo.
func logoModel(f *dra.Factory, d distro.Distro, logo distro.Logo) dra.Model {
	palette := make([]dra.Style, len(d.Colors))
	for i, c := range d.Colors {
		palette[i] = dra.Foreground(c)
	}
	var text strings.Builder
	var runs []dra.Run
	for i, line := range logo.Lines() {
		if i > 0 {
			text.WriteByte('\n')
			runs = append(runs, dra.Run{N: 1, Index: dra.NoStyle})
		}
		for _, r := range line {
			text.WriteString(r.Text)
			runs = append(runs, dra.Run{N: utf8.RuneCountInString(r.Text), Index: int8(r.Color)})
		}
	}
	m, err := f.New(dra.WithText(text.String()), dra.WithPalette(palette...), dra.WithColormap(dra.Runs(runs...)))
	if err != nil {
		// The table tests make sure that every logo fits its palette;
		// without colors the logo is still better than none.
		return f.Text(text.String())
	}
	return m
}

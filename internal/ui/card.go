package ui

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/sysinfo"
)

// group is a group of the configuration with its rows built.
type group struct {
	title string
	icon  rune
	rows  []row
}

// buildGroups builds the rows of every configured group and applies the
// label and icon settings of the items. Groups without rows are left out.
func buildGroups(info sysinfo.Info, env itemEnv) []group {
	var groups []group
	for _, g := range env.cfg.Groups {
		var rows []row
		for _, id := range g.Items {
			it, ok := registry[id]
			if !ok {
				continue
			}
			settings := env.cfg.Items[id]
			for _, r := range it.rows(info, env) {
				if settings.Label != nil {
					r.label = *settings.Label
				}
				if settings.Icon != nil {
					r.icon = firstRune(*settings.Icon)
				}
				rows = append(rows, r)
			}
		}
		if len(rows) > 0 {
			groups = append(groups, group{title: g.Title, icon: groupIcon(g), rows: rows})
		}
	}
	return groups
}

// groupIcon returns the icon of a configured group.
func groupIcon(g config.Group) rune {
	if g.Icon == config.IconAuto {
		return 0
	}
	return firstRune(g.Icon)
}

// firstRune returns the first rune of s, or 0 for an empty string.
func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

// caption returns "user@host", the known part alone when the other is
// not known, or "" when neither is.
func caption(info sysinfo.Info) string {
	var user, host string
	if info.User.State == sysinfo.Present {
		user = oneLine(info.User.V.Name)
	}
	if info.Hostname.State == sysinfo.Present {
		host = oneLine(info.Hostname.V)
	}
	switch {
	case user != "" && host != "":
		return user + "@" + host
	case user != "":
		return user
	}
	return host
}

// borderSets maps the configured border styles to the characters of
// beautify.
var borderSets = map[config.BorderStyle]blocks.BorderSet{
	config.BorderRounded: blocks.BorderRounded,
	config.BorderSquare:  blocks.BorderSquare,
	config.BorderDouble:  blocks.BorderDouble,
	config.BorderHeavy:   blocks.BorderHeavy,
	config.BorderASCII:   blocks.BorderASCII,
}

// card is the frame with one section per group.
type card struct {
	groups  []group
	theme   theme
	layout  layout
	border  config.BorderStyle
	caption string
	// dots puts the palette into the bottom edge.
	dots bool
}

// Render implements beautify.Composable. A card without groups renders
// nothing.
func (c card) Render(ctx beautify.Context) (dra.Model, error) {
	if len(c.groups) == 0 {
		return dra.Model{}, nil
	}
	all := make([][]row, len(c.groups))
	for i, g := range c.groups {
		all[i] = g.rows
	}
	labels := labelWidth(ctx, all)
	sections := make([]blocks.Section, len(c.groups))
	for i, g := range c.groups {
		title := oneLine(g.title)
		if c.layout.icons && g.icon != 0 {
			title = string(g.icon) + " " + title
		}
		rows := g.rows
		sections[i] = blocks.Section{
			Title: title,
			Body: beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
				lines := renderRows(ctx, c.theme, c.layout, rows, labels)
				return blocks.Model(dra.Stack(dra.AlignLeft, lines...)).Render(ctx)
			}),
		}
	}
	set, ok := borderSets[c.border]
	if !ok {
		set = blocks.BorderRounded
	}
	opts := []blocks.SectionsOption{
		blocks.SectionsBorder(set),
		blocks.SectionsBorderStyle(c.theme.frame),
		blocks.SectionsTitleStyle(c.theme.title),
		blocks.SectionsPadding(1, 0),
	}
	if c.caption != "" {
		opts = append(opts, blocks.SectionsCaption(c.caption), blocks.SectionsCaptionStyle(c.theme.caption))
	}
	if c.dots {
		opts = append(opts, blocks.SectionsFooter(paletteDots(ctx.Factory())))
	}
	return blocks.Sections(sections, opts...).Render(ctx)
}

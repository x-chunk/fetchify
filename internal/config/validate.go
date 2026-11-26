package config

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Limits of numeric settings.
const (
	minBarWidth = 4
	maxBarWidth = 40
	minTimeout  = time.Millisecond
	maxTimeout  = 10 * time.Second
)

// itemKeys maps the item-specific keys to the one item that accepts them.
var itemKeys = map[string]string{
	"ignore":    "disks",
	"show_type": "memory",
	"url":       "public_ip",
}

// colorPattern matches "#rgb" and "#rrggbb".
var colorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// validate checks the values of a decoded configuration that the types
// alone do not restrict. raw is the generic content of the file, which
// tells which keys were set.
func (d *decoder) validate(cfg Config, raw map[string]any) {
	d.oneOf("logo.size", string(cfg.Logo.Size), SizeAuto, SizeFull, SizeSmall)
	d.oneOf("logo.direction", string(cfg.Logo.Direction), Vertical, Horizontal, Diagonal)
	d.oneOf("colors.mode", string(cfg.Colors.Mode), ModeAuto, ModeTrueColor, Mode256, Mode16, ModeNone)
	d.oneOf("border.style", string(cfg.Border.Style), BorderRounded, BorderSquare, BorderDouble, BorderHeavy, BorderASCII)
	d.oneOf("border.color", string(cfg.Border.Color), BorderAccent, BorderLogo, BorderRainbow, BorderCustom)
	d.oneOf("border.direction", string(cfg.Border.Direction), Vertical, Horizontal, Diagonal)
	d.oneOf("bars.style", string(cfg.Bars.Style), BarBlock, BarThin)
	d.oneOf("palette.style", string(cfg.Palette.Style), PaletteDots, PaletteBlocks, PaletteNone)

	if cfg.Logo.Name != LogoAuto && cfg.Logo.Name != LogoNone && !slices.Contains(d.known.Logos, cfg.Logo.Name) {
		d.problemf("logo.name", "logo.name: unknown logo %q; use \"auto\", \"none\" or one of: %s",
			cfg.Logo.Name, strings.Join(d.known.Logos, ", "))
	}
	if cfg.Colors.Accent != AccentAuto {
		d.color("colors.accent", cfg.Colors.Accent)
	}
	for _, c := range cfg.Logo.Colors {
		d.color("logo.colors", c)
	}
	for _, c := range cfg.Border.Colors {
		d.color("border.colors", c)
	}
	if cfg.Border.Color == BorderCustom && len(cfg.Border.Colors) == 0 {
		key := "border.colors"
		if _, ok := d.positions[key]; !ok {
			key = "border.color"
		}
		d.problemf(key, "border.colors: border.color = \"custom\" needs at least one color")
	}
	if w := cfg.Bars.Width; w < minBarWidth || w > maxBarWidth {
		d.problemf("bars.width", "bars.width: %d is outside %d-%d", w, minBarWidth, maxBarWidth)
	}
	if th := cfg.Bars.Thresholds; th[0] < 0 || th[1] > 100 || th[0] >= th[1] {
		d.problemf("bars.thresholds", "bars.thresholds: [%d, %d] must be two ascending numbers within 0-100", th[0], th[1])
	}
	d.timeoutRange("timeout_ms", cfg.Timeout)

	d.validateGroups(cfg.Groups)
	d.validateItems(cfg.Items, raw)
}

// validateGroups checks titles, item identifiers and duplicates.
func (d *decoder) validateGroups(groups []Group) {
	where := map[string]string{} // item -> title of the group that lists it
	for i, g := range groups {
		key := "group." + strconv.Itoa(i)
		if strings.TrimSpace(g.Title) == "" {
			d.problemf(key, "group %d: the title is empty", i+1)
		}
		for _, id := range g.Items {
			if !slices.Contains(d.known.Items, id) {
				msg := fmt.Sprintf("%s.items: unknown item %q", key, id)
				if near := closest(id, d.known.Items); near != "" {
					msg += fmt.Sprintf("; did you mean %q?", near)
				}
				d.problemf(key+".items", "%s", msg)
				continue
			}
			if first, ok := where[id]; ok {
				d.problemf(key+".items", "%s.items: item %q is listed in group %q and in group %q", key, id, first, g.Title)
				continue
			}
			where[id] = g.Title
		}
	}
}

// validateItems checks the [items.<id>] tables.
func (d *decoder) validateItems(items map[string]Item, raw map[string]any) {
	tables, _ := raw["items"].(map[string]any)
	for _, id := range slices.Sorted(maps.Keys(tables)) {
		key := "items." + id
		if !slices.Contains(d.known.Items, id) {
			d.problemf(key, "%s: unknown item %q", key, id)
			continue
		}
		t, _ := tables[id].(map[string]any)
		for k, owner := range itemKeys {
			if _, ok := t[k]; ok && owner != id {
				d.problemf(key+"."+k, "%s.%s: %s is only valid for item %s", key, k, k, owner)
			}
		}
		if _, ok := t["timeout_ms"]; ok {
			d.timeoutRange(key+".timeout_ms", items[id].Timeout)
		}
	}
}

// oneOf records a value outside its enumeration.
func (d *decoder) oneOf(key, value string, allowed ...any) {
	names := make([]string, len(allowed))
	for i, a := range allowed {
		names[i] = fmt.Sprint(a)
		if names[i] == value {
			return
		}
	}
	d.problemf(key, "%s: %q is not one of %s", key, value, strings.Join(names, ", "))
}

// color records a color that is not "#rgb" or "#rrggbb".
func (d *decoder) color(key, value string) {
	if !colorPattern.MatchString(value) {
		d.problemf(key, "%s: %q is not a color of the form #rgb or #rrggbb", key, value)
	}
}

// timeoutRange records a time limit outside 1-10000 ms.
func (d *decoder) timeoutRange(key string, t time.Duration) {
	if t < minTimeout || t > maxTimeout {
		d.problemf(key, "%s: %d is outside %d-%d", key, t.Milliseconds(), minTimeout.Milliseconds(), maxTimeout.Milliseconds())
	}
}

// closest returns the candidate nearest to s by edit distance when it is
// at most 2, or "".
func closest(s string, candidates []string) string {
	best, bestDist := "", 3
	for _, c := range candidates {
		if dist := editDistance(s, c); dist < bestDist {
			best, bestDist = c, dist
		}
	}
	return best
}

// editDistance returns the Levenshtein distance between a and b.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

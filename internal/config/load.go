package config

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// Known holds the identifiers a configuration may refer to. The caller
// provides them so that config does not depend on the presentation
// layer.
type Known struct {
	// Items are the identifiers of all items, default and opt-in.
	Items []string
	// Logos are the identifiers of all logos.
	Logos []string
}

// DefaultPath returns the location of the configuration file:
// $XDG_CONFIG_HOME/fetchify/config.toml, or ~/.config/fetchify/config.toml
// when the variable is unset or not absolute. It returns "" when neither
// is known, for example without a home directory.
func DefaultPath(getenv func(string) string, home string) string {
	if dir := getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "fetchify", "config.toml")
	}
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "fetchify", "config.toml")
}

// Load reads the configuration file at path and merges it over the
// defaults. A file that cannot be read returns the error of the file
// system, so callers can tell a missing file with errors.Is and
// fs.ErrNotExist. Invalid content returns an *Error that lists every
// problem found.
func Load(path string, known Known) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return Parse(path, data, known)
}

// Parse is Load for content that is already in memory. path is only used
// in error messages.
func Parse(path string, data []byte, known Known) (Config, error) {
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		var de *toml.DecodeError
		if errors.As(err, &de) {
			line, col := de.Position()
			msg := strings.TrimPrefix(de.Error(), "toml: ")
			return Config{}, &Error{Path: path, Problems: []Problem{{Line: line, Column: col, Message: msg}}}
		}
		return Config{}, &Error{Path: path, Problems: []Problem{{Message: err.Error()}}}
	}
	d := &decoder{positions: keyPositions(data), known: known}
	cfg := d.config(raw)
	// A value of the wrong type was not stored and the default stays, so
	// validation does not report the same mistake twice.
	d.validate(cfg, raw)
	if len(d.problems) > 0 {
		slices.SortStableFunc(d.problems, func(a, b Problem) int {
			if (a.Line == 0) != (b.Line == 0) {
				return cmp.Compare(b.Line, a.Line) // unknown positions last
			}
			return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column))
		})
		return Config{}, &Error{Path: path, Problems: d.problems}
	}
	return cfg, nil
}

// keyPositions maps every key of a TOML document to the position of its
// first appearance. Keys are dotted paths; entries of an array of tables
// carry their index, as in "group.1.title".
func keyPositions(data []byte) map[string][2]int {
	positions := map[string][2]int{}
	counts := map[string]int{}
	var table []string
	var p unstable.Parser
	p.Reset(data)
	for p.NextExpression() {
		e := p.Expression()
		var key []string
		var first *unstable.Node
		for it := e.Key(); it.Next(); {
			n := it.Node()
			if first == nil {
				first = n
			}
			key = append(key, string(n.Data))
		}
		if first == nil {
			continue
		}
		shape := p.Shape(first.Raw)
		pos := [2]int{shape.Start.Line, shape.Start.Column}
		var full []string
		switch e.Kind {
		case unstable.Table:
			table = key
			full = key
		case unstable.ArrayTable:
			name := strings.Join(key, ".")
			table = append(slices.Clone(key), strconv.Itoa(counts[name]))
			counts[name]++
			full = table
		case unstable.KeyValue:
			full = append(slices.Clone(table), key...)
		default:
			continue
		}
		for i := range full {
			k := strings.Join(full[:i+1], ".")
			if _, ok := positions[k]; !ok {
				positions[k] = pos
			}
		}
	}
	return positions
}

// decoder turns the generic content of a TOML file into a Config and
// collects every problem on the way.
type decoder struct {
	positions map[string][2]int
	known     Known
	problems  []Problem
}

// problemf records a problem at the position of key.
func (d *decoder) problemf(key, format string, args ...any) {
	pos := d.positions[key]
	d.problems = append(d.problems, Problem{Line: pos[0], Column: pos[1], Message: fmt.Sprintf(format, args...)})
}

// config decodes the whole file over the defaults.
func (d *decoder) config(raw map[string]any) Config {
	cfg := Defaults()
	d.unknown(raw, "", "clear", "footer", "timeout_ms", "logo", "colors", "border", "icons", "bars", "palette", "group", "items")
	d.boolean(raw, "clear", &cfg.Clear)
	d.boolean(raw, "footer", &cfg.Footer)
	d.timeout(raw, "timeout_ms", &cfg.Timeout)
	if t, ok := d.table(raw, "logo"); ok {
		d.unknown(t, "logo", "name", "size", "gradient", "colors", "direction")
		d.str(t, "logo.name", &cfg.Logo.Name)
		d.str(t, "logo.size", (*string)(&cfg.Logo.Size))
		d.boolean(t, "logo.gradient", &cfg.Logo.Gradient)
		d.stringList(t, "logo.colors", &cfg.Logo.Colors)
		d.str(t, "logo.direction", (*string)(&cfg.Logo.Direction))
	}
	if t, ok := d.table(raw, "colors"); ok {
		d.unknown(t, "colors", "mode", "accent")
		d.str(t, "colors.mode", (*string)(&cfg.Colors.Mode))
		d.str(t, "colors.accent", &cfg.Colors.Accent)
	}
	if t, ok := d.table(raw, "border"); ok {
		d.unknown(t, "border", "style", "color", "colors", "direction")
		d.str(t, "border.style", (*string)(&cfg.Border.Style))
		d.str(t, "border.color", (*string)(&cfg.Border.Color))
		d.stringList(t, "border.colors", &cfg.Border.Colors)
		d.str(t, "border.direction", (*string)(&cfg.Border.Direction))
	}
	if t, ok := d.table(raw, "icons"); ok {
		d.unknown(t, "icons", "enabled")
		d.boolean(t, "icons.enabled", &cfg.Icons.Enabled)
	}
	if t, ok := d.table(raw, "bars"); ok {
		d.unknown(t, "bars", "style", "width", "thresholds")
		d.str(t, "bars.style", (*string)(&cfg.Bars.Style))
		d.integer(t, "bars.width", &cfg.Bars.Width)
		d.thresholds(t, "bars.thresholds", &cfg.Bars.Thresholds)
	}
	if t, ok := d.table(raw, "palette"); ok {
		d.unknown(t, "palette", "style")
		d.str(t, "palette.style", (*string)(&cfg.Palette.Style))
	}
	if v, ok := raw["group"]; ok {
		cfg.Groups = d.groups(v)
	}
	if t, ok := d.table(raw, "items"); ok {
		d.items(t, cfg.Items)
	}
	return cfg
}

// groups decodes the array of [[group]] tables. Defining any group
// replaces the default groups.
func (d *decoder) groups(v any) []Group {
	list, ok := v.([]any)
	if !ok {
		d.problemf("group", "group: expected an array of tables, got %s", typeName(v))
		return nil
	}
	groups := []Group{}
	for i, entry := range list {
		key := "group." + strconv.Itoa(i)
		t, ok := entry.(map[string]any)
		if !ok {
			d.problemf("group", "group %d: expected a table, got %s", i+1, typeName(entry))
			continue
		}
		g := Group{Icon: IconAuto, Items: []string{}}
		d.unknown(t, key, "title", "icon", "items")
		d.str(t, key+".title", &g.Title)
		d.str(t, key+".icon", &g.Icon)
		d.stringList(t, key+".items", &g.Items)
		groups = append(groups, g)
	}
	return groups
}

// items decodes the [items.<id>] tables into settings.
func (d *decoder) items(t map[string]any, settings map[string]Item) {
	for _, id := range slices.Sorted(maps.Keys(t)) {
		key := "items." + id
		it, ok := d.table(t, id, key)
		if !ok {
			continue
		}
		item := settings[id]
		d.unknown(it, key, "label", "icon", "timeout_ms", "ignore", "show_type", "url")
		if v, ok := it["label"]; ok {
			if s, ok := v.(string); ok {
				item.Label = &s
			} else {
				d.typeProblem(key+".label", "a string", v)
			}
		}
		if v, ok := it["icon"]; ok {
			if s, ok := v.(string); ok {
				item.Icon = &s
			} else {
				d.typeProblem(key+".icon", "a string", v)
			}
		}
		d.timeout(it, key+".timeout_ms", &item.Timeout)
		d.stringList(it, key+".ignore", &item.Ignore)
		d.boolean(it, key+".show_type", &item.ShowType)
		d.str(it, key+".url", &item.URL)
		settings[id] = item
	}
}

// last returns the last component of a dotted key.
func last(key string) string {
	return key[strings.LastIndexByte(key, '.')+1:]
}

// unknown records every key of t that is not in allowed. prefix is the
// dotted path of t.
func (d *decoder) unknown(t map[string]any, prefix string, allowed ...string) {
	for _, k := range slices.Sorted(maps.Keys(t)) {
		if slices.Contains(allowed, k) {
			continue
		}
		full := k
		if prefix != "" {
			full = prefix + "." + k
		}
		d.problemf(full, "unknown key %s", full)
	}
}

// table returns the subtable at name of t. key is its full dotted path
// and defaults to name.
func (d *decoder) table(t map[string]any, name string, key ...string) (map[string]any, bool) {
	full := name
	if len(key) > 0 {
		full = key[0]
	}
	v, ok := t[name]
	if !ok {
		return nil, false
	}
	sub, ok := v.(map[string]any)
	if !ok {
		d.typeProblem(full, "a table", v)
		return nil, false
	}
	return sub, true
}

// typeProblem records a value of the wrong type.
func (d *decoder) typeProblem(key, want string, got any) {
	d.problemf(key, "%s: expected %s, got %s", key, want, typeName(got))
}

func (d *decoder) boolean(t map[string]any, key string, dst *bool) {
	if v, ok := t[last(key)]; ok {
		if b, ok := v.(bool); ok {
			*dst = b
		} else {
			d.typeProblem(key, "a boolean", v)
		}
	}
}

func (d *decoder) str(t map[string]any, key string, dst *string) {
	if v, ok := t[last(key)]; ok {
		if s, ok := v.(string); ok {
			*dst = s
		} else {
			d.typeProblem(key, "a string", v)
		}
	}
}

func (d *decoder) integer(t map[string]any, key string, dst *int) {
	if v, ok := t[last(key)]; ok {
		if n, ok := v.(int64); ok {
			*dst = int(n)
		} else {
			d.typeProblem(key, "an integer", v)
		}
	}
}

func (d *decoder) stringList(t map[string]any, key string, dst *[]string) {
	v, ok := t[last(key)]
	if !ok {
		return
	}
	list, ok := v.([]any)
	if !ok {
		d.typeProblem(key, "an array of strings", v)
		return
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		s, ok := e.(string)
		if !ok {
			d.typeProblem(key, "an array of strings", e)
			return
		}
		out = append(out, s)
	}
	*dst = out
}

// timeout decodes a time limit given in milliseconds.
func (d *decoder) timeout(t map[string]any, key string, dst *time.Duration) {
	var ms int
	if _, ok := t[last(key)]; !ok {
		return
	}
	before := len(d.problems)
	d.integer(t, key, &ms)
	if len(d.problems) == before {
		*dst = time.Duration(ms) * time.Millisecond
	}
}

// thresholds decodes bars.thresholds.
func (d *decoder) thresholds(t map[string]any, key string, dst *[2]int) {
	v, ok := t[last(key)]
	if !ok {
		return
	}
	list, ok := v.([]any)
	if !ok {
		d.typeProblem(key, "an array of two integers", v)
		return
	}
	var out [2]int
	for i, e := range list {
		n, ok := e.(int64)
		if !ok || i >= 2 {
			d.problemf(key, "%s: expected an array of two integers", key)
			return
		}
		out[i] = int(n)
	}
	if len(list) != 2 {
		d.problemf(key, "%s: expected an array of two integers", key)
		return
	}
	*dst = out
}

// typeName names the TOML type of a decoded value.
func typeName(v any) string {
	switch v.(type) {
	case string:
		return "a string"
	case bool:
		return "a boolean"
	case int64:
		return "an integer"
	case float64:
		return "a float"
	case []any:
		return "an array"
	case map[string]any:
		return "a table"
	case toml.LocalDate, toml.LocalTime, toml.LocalDateTime, time.Time:
		return "a date or time"
	}
	return fmt.Sprintf("%T", v)
}

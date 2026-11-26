package distro

import (
	"embed"
	"maps"
	"slices"

	"go.xchunk.org/beautify/dra"
)

// Distro is what fetchify knows about a distribution.
type Distro struct {
	// ID is the lookup key in lower case, as in the ID of os-release.
	ID string
	// Name is the display name, used when os-release has none.
	Name string
	// Colors is the palette of the logo, in the order of its markers.
	// Entries may be dra.TerminalDefault for parts drawn in the
	// terminal's own foreground.
	Colors []dra.Color
	// Accent is the color of the frame, the titles, the icons and the
	// labels.
	Accent dra.Color
	// Icon is the Nerd Font glyph of the distribution.
	Icon rune
}

// Generic entries of the table.
const (
	// LinuxID is Tux, the fallback for unknown distributions.
	LinuxID = "linux"
	// KoalaID is the fetchify mascot. Lookup never chooses it.
	KoalaID = "koala"
)

//go:embed logos/*.txt
var logoFiles embed.FS

// table holds every distribution by ID.
var table = map[string]Distro{}

// aliases maps further identifiers to entries of table.
var aliases = map[string]string{}

// register adds entries to the table. It is called from init functions of
// the files that hold the entries.
func register(entries ...Distro) {
	for _, d := range entries {
		table[d.ID] = d
	}
}

// alias makes the identifiers ids resolve to the entry id.
func alias(id string, ids ...string) {
	for _, a := range ids {
		aliases[a] = id
	}
}

// hex returns the color written as "#rrggbb". It panics on a malformed
// value, which only a typo in the table can cause.
func hex(s string) dra.Color {
	c, err := dra.Hex(s)
	if err != nil {
		panic(err)
	}
	return c
}

// ByID returns the entry with the identifier id or an alias of it, and
// reports whether there is one.
func ByID(id string) (Distro, bool) {
	if target, ok := aliases[id]; ok {
		id = target
	}
	d, ok := table[id]
	if !ok {
		return Distro{}, false
	}
	d.Colors = slices.Clone(d.Colors)
	return d, true
}

// IDs returns the identifiers of all entries, sorted. Aliases are not
// included.
func IDs() []string {
	return slices.Sorted(maps.Keys(table))
}

// Lookup returns the entry for a release: the one of its ID, then the
// first one of its IDLike list, so an unknown derivative gets the logo of
// its parent, and otherwise Tux. It never fails and never returns the
// koala.
func Lookup(r Release) Distro {
	for _, id := range append([]string{r.ID}, r.IDLike...) {
		if id == KoalaID {
			continue
		}
		if d, ok := ByID(id); ok {
			return d
		}
	}
	d, _ := ByID(LinuxID)
	return d
}

// Logo returns the logo of the distribution in the given size, parsed
// against its palette. An entry without a logo file has an empty logo;
// the tests make sure that no entry of the table lacks one.
func (d Distro) Logo(size Size) Logo {
	name := "logos/" + d.ID + ".txt"
	if size == Small {
		name = "logos/" + d.ID + "_small.txt"
	}
	data, err := logoFiles.ReadFile(name)
	if err != nil {
		return Logo{}
	}
	return ParseLogo(string(data), len(d.Colors))
}

func init() {
	register(
		Distro{
			ID:     LinuxID,
			Name:   "Linux",
			Colors: []dra.Color{dra.TerminalDefault, hex("#7f7f7f"), hex("#f0c000")},
			Accent: hex("#f0c000"),
			Icon:   '', // nf-linux-tux
		},
		Distro{
			ID:     KoalaID,
			Name:   "fetchify",
			Colors: []dra.Color{hex("#9e9e9e"), hex("#e8a0b4")},
			Accent: hex("#8fbf8f"),
			Icon:   '\U000f173f', // nf-md-koala
		},
	)
}

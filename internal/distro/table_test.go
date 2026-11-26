package distro

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"go.xchunk.org/beautify/dra"
)

func TestTableIntegrity(t *testing.T) {
	for _, id := range IDs() {
		d, _ := ByID(id)
		if d.ID != id || id != strings.ToLower(id) {
			t.Errorf("%s: entry has ID %q", id, d.ID)
		}
		if d.Name == "" || len(d.Colors) == 0 || !d.Accent.IsSet() || d.Accent == dra.TerminalDefault {
			t.Errorf("%s: incomplete entry %+v", id, d)
		}
		if !unicode.Is(unicode.Co, d.Icon) {
			t.Errorf("%s: icon %U is not in a private-use range", id, d.Icon)
		}
		for _, size := range []Size{Full, Small} {
			if d.Logo(size).Height() == 0 {
				t.Errorf("%s: logo of size %d is missing", id, size)
			}
		}
	}
	for a, id := range aliases {
		if _, ok := table[id]; !ok {
			t.Errorf("alias %s points to unknown %s", a, id)
		}
		if _, ok := table[a]; ok {
			t.Errorf("alias %s shadows an entry", a)
		}
	}
}

// marker matches the color markers of the logo format.
var marker = regexp.MustCompile(`\$\{([1-9])\}`)

func TestLogoFiles(t *testing.T) {
	files, err := fs.Glob(logoFiles, "logos/*.txt")
	if err != nil || len(files) == 0 {
		t.Fatalf("no logo files: %v", err)
	}
	for _, name := range files {
		id := strings.TrimSuffix(strings.TrimSuffix(path.Base(name), ".txt"), "_small")
		d, ok := table[id]
		if !ok {
			t.Errorf("%s: no table entry %q", name, id)
			continue
		}
		data, _ := logoFiles.ReadFile(name)
		text := string(data)
		if !strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\n") {
			t.Errorf("%s: must end with exactly one newline", name)
		}
		for i, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
			if strings.HasSuffix(line, " ") {
				t.Errorf("%s:%d: trailing space", name, i+1)
			}
			for _, r := range line {
				if r == '\t' || unicode.IsControl(r) {
					t.Errorf("%s:%d: control character %U", name, i+1, r)
				}
			}
			for _, m := range marker.FindAllStringSubmatch(line, -1) {
				if int(m[1][0]-'0') > len(d.Colors) {
					t.Errorf("%s:%d: marker %s refers to a missing color", name, i+1, m[0])
				}
			}
		}
		logo := ParseLogo(text, len(d.Colors))
		maxW, maxH := 40, 20
		if strings.HasSuffix(name, "_small.txt") {
			maxW, maxH = 16, 8
		}
		if logo.Width() > maxW || logo.Height() > maxH {
			t.Errorf("%s: %dx%d exceeds %dx%d", name, logo.Width(), logo.Height(), maxW, maxH)
		}
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		name string
		r    Release
		want string
	}{
		{"direct", Release{ID: "arch"}, "arch"},
		{"alias", Release{ID: "archarm"}, "arch"},
		{"ID_LIKE chain", Release{ID: "garuda", IDLike: []string{"nope", "arch"}}, "arch"},
		{"derivative", Release{ID: "neon", IDLike: []string{"ubuntu", "debian"}}, "ubuntu"},
		{"raspbian", Release{ID: "raspbian", IDLike: []string{"debian"}}, "debian"},
		{"opensuse alias", Release{ID: "opensuse-tumbleweed", IDLike: []string{"opensuse", "suse"}}, "opensuse"},
		{"unknown", Release{ID: "plan9"}, LinuxID},
		{"empty", Release{}, LinuxID},
		{"koala is never chosen", Release{ID: "koala"}, LinuxID},
	}
	for _, tt := range tests {
		if got := Lookup(tt.r).ID; got != tt.want {
			t.Errorf("%s: Lookup = %q, want %q", tt.name, got, tt.want)
		}
	}
	if _, ok := ByID("nope"); ok {
		t.Error("ByID found an unknown ID")
	}
	d, _ := ByID(LinuxID)
	d.Colors[0] = dra.Red
	if again, _ := ByID(LinuxID); again.Colors[0] == dra.Red {
		t.Error("ByID shares the palette with the table")
	}
}

// TestTableSize pins the first version of the table: 22 distributions
// and the two generic entries.
func TestTableSize(t *testing.T) {
	if n := len(IDs()); n != 24 {
		t.Errorf("table has %d entries, want 24: %v", n, IDs())
	}
}

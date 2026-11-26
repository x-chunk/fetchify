package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"enumeration", "[border]\nstyle = \"dotted\"",
			[]string{`2:1: border.style: "dotted" is not one of rounded, square, double, heavy, ascii`}},
		{"unknown item with suggestion", "[[group]]\ntitle = \"A\"\nitems = [\"kernl\"]",
			[]string{`3:1: group.0.items: unknown item "kernl"; did you mean "kernel"?`}},
		{"unknown item without suggestion", "[[group]]\ntitle = \"A\"\nitems = [\"weather\"]",
			[]string{`3:1: group.0.items: unknown item "weather"`}},
		{"twice in one group", "[[group]]\ntitle = \"A\"\nitems = [\"os\", \"os\"]",
			[]string{`3:1: group.0.items: item "os" is listed in group "A" and in group "A"`}},
		{"twice in two groups", "[[group]]\ntitle = \"A\"\nitems = [\"os\"]\n[[group]]\ntitle = \"B\"\nitems = [\"os\"]",
			[]string{`6:1: group.1.items: item "os" is listed in group "A" and in group "B"`}},
		{"settings of an unknown item", "[items.weather]\nlabel = \"W\"",
			[]string{`1:2: items.weather: unknown item "weather"`}},
		{"key of another item", "[items.cpu]\nignore = [\"/x\"]",
			[]string{`2:1: items.cpu.ignore: ignore is only valid for item disks`}},
		{"color", "[colors]\naccent = \"blue\"\n[logo]\ncolors = [\"#12\"]",
			[]string{`2:1: colors.accent: "blue" is not a color of the form #rgb or #rrggbb`,
				`4:1: logo.colors: "#12" is not a color of the form #rgb or #rrggbb`}},
		{"custom border without colors", "[border]\ncolor = \"custom\"",
			[]string{`2:1: border.colors: border.color = "custom" needs at least one color`}},
		{"bar width", "[bars]\nwidth = 3", []string{`2:1: bars.width: 3 is outside 4-40`}},
		{"thresholds", "[bars]\nthresholds = [90, 80]",
			[]string{`2:1: bars.thresholds: [90, 80] must be two ascending numbers within 0-100`}},
		{"thresholds above 100", "[bars]\nthresholds = [50, 101]",
			[]string{`2:1: bars.thresholds: [50, 101] must be two ascending numbers within 0-100`}},
		{"timeouts", "timeout_ms = 0\n[items.cpu]\ntimeout_ms = 20000",
			[]string{`1:1: timeout_ms: 0 is outside 1-10000`, `3:1: items.cpu.timeout_ms: 20000 is outside 1-10000`}},
		{"logo name", "[logo]\nname = \"windows\"",
			[]string{`2:1: logo.name: unknown logo "windows"; use "auto", "none" or one of: arch, ubuntu, koala, linux`}},
		{"empty title", "[[group]]\nitems = []", []string{`1:3: group 1: the title is empty`}},
		{"several at once", "colour = 1\n[bars]\nwidth = 99\n[palette]\nstyle = 1",
			[]string{"1:1: unknown key colour", "3:1: bars.width: 99 is outside 4-40",
				"5:1: palette.style: expected a string, got an integer"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse(t, tt.text)
			if err == nil {
				t.Fatal("no error")
			}
			if got := problems(t, err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestValidConfigurations(t *testing.T) {
	for _, text := range []string{
		"[[group]]\ntitle = \"Empty\"\nitems = []",
		"[logo]\nname = \"koala\"\n[colors]\naccent = \"#ABCDEF\"",
		"[border]\ncolor = \"custom\"\ncolors = [\"#f00\"]",
		"[items.disks]\nignore = []\n[items.memory]\nshow_type = true\n[items.public_ip]\nurl = \"https://example.com\"",
	} {
		if _, err := parse(t, text); err != nil {
			t.Errorf("%q: %v", text, err)
		}
	}
}

func TestEditDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{{"", "", 0}, {"kernl", "kernel", 1}, {"gpu", "cpu", 1}, {"abc", "", 3}, {"swap", "wsap", 2}}
	for _, tt := range tests {
		if got := editDistance(tt.a, tt.b); got != tt.want {
			t.Errorf("editDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func FuzzLoad(f *testing.F) {
	f.Add(DefaultTOML())
	f.Add([]byte("[[group]]\ntitle = 1\nitems = [{a = 1}]\n[items.x.y]\nz = 1979-05-27"))
	f.Fuzz(func(t *testing.T, data []byte) {
		Parse("fuzz.toml", data, known)
	})
}

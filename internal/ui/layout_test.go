package ui

import (
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
)

// testTheme is the theme of an Arch machine with the default
// configuration.
func testTheme() theme {
	arch, _ := distro.ByID("arch")
	return newTheme(config.Defaults(), arch, false)
}

// plainLines renders lines without styles.
func plainLines(lines []dra.Model) string {
	return dra.NewEngine(dra.ProfilePlain).String(dra.Stack(dra.AlignLeft, lines...))
}

func TestRenderRowsAlignsLabels(t *testing.T) {
	ctx := beautify.NewContext(nil)
	system := []row{
		{icon: 'A', label: "OS", value: textValue("Arch Linux")},
		{label: "Kernel", value: []segment{{text: "6.10"}, {text: " x", role: roleMuted}}},
	}
	storage := []row{{icon: 'B', label: "Packages", value: textValue("1")}}
	width := labelWidth(ctx, [][]row{system, storage})
	if width != 8 {
		t.Fatalf("label width = %d, want 8", width)
	}
	got := plainLines(renderRows(ctx, testTheme(), layout{icons: true}, system, width))
	want := "A OS        Arch Linux\n  Kernel    6.10 x"
	if got != want {
		t.Errorf("with icons:\n%s\nwant\n%s", got, want)
	}
	got = plainLines(renderRows(ctx, testTheme(), layout{}, system, width))
	want = "OS        Arch Linux\nKernel    6.10 x"
	if got != want {
		t.Errorf("without icons:\n%s\nwant\n%s", got, want)
	}
}

func TestRenderRowsStyles(t *testing.T) {
	ctx := beautify.NewContext(nil)
	rows := []row{{label: "K", value: []segment{{text: "v"}, {text: "m", role: roleMuted}}}}
	got := dra.NewEngine(dra.ProfileANSI16).String(dra.Stack(dra.AlignLeft, renderRows(ctx, testTheme(), layout{}, rows, 1)...))
	if !strings.Contains(got, "\x1b[0;90mm") {
		t.Errorf("muted segment is not bright black: %q", got)
	}
	if strings.Contains(got, "\n") || !strings.HasPrefix(got, "\x1b[0;") {
		t.Errorf("unexpected rendering %q", got)
	}
}

func TestRenderRowsHostileText(t *testing.T) {
	ctx := beautify.NewContext(nil)
	rows := []row{{label: "a\nb", value: textValue("x\x1b[2Jy\nz")}}
	got := plainLines(renderRows(ctx, testTheme(), layout{}, rows, 3))
	if strings.Count(got, "\n") != 0 || strings.ContainsRune(got, '\x1b') {
		t.Errorf("hostile text broke the row: %q", got)
	}
}

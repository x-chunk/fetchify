package ui

import (
	"fmt"
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/golden"
)

func TestResponsiveSteps(t *testing.T) {
	s := newScreen(Input{Info: ubuntuDesktop(), Config: config.Defaults()})
	tests := []struct {
		width int // 0 is unbounded
		want  step
	}{
		{0, stepFullLogo},
		{200, stepFullLogo}, {120, stepFullLogo}, {102, stepFullLogo},
		{100, stepSmallLogo}, {80, stepSmallLogo},
		{60, stepNoLogo},
		{55, stepNoDetails},
		{50, stepNarrowBars}, {45, stepNarrowBars},
		{42, stepNoBars},
		{40, stepTruncated}, {20, stepTruncated}, {10, stepTruncated}, {1, stepTruncated},
	}
	for _, tt := range tests {
		ctx := beautify.NewContext(nil)
		if tt.width > 0 {
			ctx = ctx.WithWidth(tt.width)
		}
		m, got, err := s.render(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("width %d: step %d, want %d (rendered %d wide)", tt.width, got, tt.want, m.Width())
		}
		if tt.width > 0 && m.Width() > tt.width {
			t.Errorf("width %d: rendered %d wide", tt.width, m.Width())
		}
	}
	m, _, _ := s.render(beautify.NewContext(nil).WithWidth(0))
	if !m.IsEmpty() {
		t.Errorf("width 0 rendered %q", m.String())
	}
}

func TestResponsiveGoldenFiles(t *testing.T) {
	for _, w := range []int{80, 60, 55, 45, 42, 40, 20} {
		t.Run(fmt.Sprint(w), func(t *testing.T) {
			plain := renderScreen(t, ubuntuDesktop(), config.Defaults(), dra.ProfilePlain, w)
			golden.Assert(t, fmt.Sprintf("responsive_%d", w), []byte(plain+"\n"))
			if w >= 40 {
				assertCardAligned(t, plain)
			}
		})
	}
}

func TestPinnedLogoSize(t *testing.T) {
	cfg := config.Defaults()
	cfg.Logo.Size = config.SizeFull
	s := newScreen(Input{Info: ubuntuDesktop(), Config: cfg})
	if _, got, _ := s.render(beautify.NewContext(nil).WithWidth(90)); got != stepNoLogo {
		t.Errorf("pinned full logo at 90: step %d, want no logo", got)
	}
	cfg.Logo.Name = config.LogoNone
	s = newScreen(Input{Info: ubuntuDesktop(), Config: cfg})
	if _, got, _ := s.render(beautify.NewContext(nil).WithWidth(200)); got != stepNoLogo {
		t.Errorf("logo none: step %d", got)
	}
}

func TestHostileValues(t *testing.T) {
	info := ubuntuDesktop()
	info.Hostname = present("box\x1b[2J")
	info.Kernel = present("6.8\nevil")
	info.Host = present(strings.Repeat("Very Long Model Name ", 20))
	info.Terminal = present("終端エミュレータ")
	plain := renderScreen(t, info, config.Defaults(), dra.ProfilePlain, 0)
	if strings.ContainsRune(plain, '\x1b') {
		t.Error("escape character reached the output")
	}
	assertCardAligned(t, plain)
	narrow := renderScreen(t, info, config.Defaults(), dra.ProfilePlain, 80)
	for _, line := range strings.Split(narrow, "\n") {
		if w := dra.StringWidth(line, dra.ModeReal); w > 80 {
			t.Errorf("line of %d cells at width 80: %q", w, line)
		}
	}
}

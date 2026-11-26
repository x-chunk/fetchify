package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// FactTerminalFont is the font of the terminal emulator.
const FactTerminalFont Fact = "terminal_font"

// terminalFontReader reads the font of one terminal from its
// configuration directory.
type terminalFontReader func(fsys fs.FS, config, home string) string

var (
	// alacrittyFamily matches the family in alacritty.toml, both in a
	// [font.normal] table and in an inline table.
	alacrittyFamily = regexp.MustCompile(`family\s*=\s*"([^"]+)"`)
	// weztermFont matches the first font name given to wezterm.font or
	// wezterm.font_with_fallback.
	weztermFont = regexp.MustCompile(`wezterm\.font(?:_with_fallback)?\s*\(\s*\{?\s*["']([^"']+)["']`)
)

// terminalFonts maps the terminals whose configuration is one known file
// to their readers.
var terminalFonts = map[string]terminalFontReader{
	"kitty": func(fsys fs.FS, config, _ string) string {
		return spaceValue(fsys, path.Join(config, "kitty/kitty.conf"), "font_family")
	},
	"Alacritty": func(fsys fs.FS, config, _ string) string {
		return firstMatch(fsys, path.Join(config, "alacritty/alacritty.toml"), alacrittyFamily)
	},
	"foot": func(fsys fs.FS, config, _ string) string {
		font := iniValue(fsys, path.Join(config, "foot/foot.ini"), "font")
		font, _, _ = strings.Cut(font, ",") // fallback fonts
		font, _, _ = strings.Cut(font, ":") // attributes such as size
		return strings.TrimSpace(font)
	},
	"WezTerm": func(fsys fs.FS, config, home string) string {
		if font := firstMatch(fsys, path.Join(config, "wezterm/wezterm.lua"), weztermFont); font != "" {
			return font
		}
		return firstMatch(fsys, path.Join(home, ".wezterm.lua"), weztermFont)
	},
	"Ghostty": func(fsys fs.FS, config, _ string) string {
		return iniValue(fsys, path.Join(config, "ghostty/config"), "font-family")
	},
}

func init() {
	collectors[FactTerminalFont] = fact(collectTerminalFont, func(i *Info) *Value[string] { return &i.TerminalFont })
}

// collectTerminalFont detects the terminal and reads the font from its
// configuration. The fact is Absent without a terminal emulator and
// Unknown for other terminals or when the configuration sets no font,
// because the terminal's built-in default is not known.
func collectTerminalFont(ctx context.Context, s *session) Value[string] {
	terminal := collectTerminal(ctx, s)
	if terminal.State != Present {
		return Value[string]{State: terminal.State}
	}
	read, ok := terminalFonts[terminal.V]
	if !ok {
		return Value[string]{}
	}
	if font := read(s.fsys(), s.configHome(), s.home()); font != "" {
		return present(font)
	}
	return Value[string]{}
}

// spaceValue returns the value of key in a configuration file of
// "key value" lines, as kitty uses, or "".
func spaceValue(fsys fs.FS, name, key string) string {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return ""
	}
	value := ""
	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == key {
			// The last setting wins, as in kitty itself.
			value = strings.Join(fields[1:], " ")
		}
	}
	return value
}

// firstMatch returns the first submatch of re in the file name, or "".
func firstMatch(fsys fs.FS, name string, re *regexp.Regexp) string {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return ""
	}
	if m := re.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return ""
}

package sysinfo

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestTerminalFont(t *testing.T) {
	files := fstest.MapFS{
		"home/u/.config/kitty/kitty.conf":         {Data: []byte("# font_family Nope\nfont_family Iosevka Term\nfont_family  JetBrains Mono\n")},
		"home/u/.config/alacritty/alacritty.toml": {Data: []byte("[font]\nsize = 11\n[font.normal]\nfamily = \"Fira Code\"\n")},
		"home/u/.config/foot/foot.ini":            {Data: []byte("[main]\nfont=Hack:size=10,Noto Color Emoji\n")},
		"home/u/.wezterm.lua":                     {Data: []byte("config.font = wezterm.font_with_fallback({ 'Cascadia Code', 'Noto' })\n")},
		"home/u/.config/ghostty/config":           {Data: []byte("font-family = Berkeley Mono\n")},
	}
	tests := []struct {
		name string
		env  map[string]string
		want Value[string]
	}{
		{"kitty", map[string]string{"KITTY_WINDOW_ID": "1"}, present("JetBrains Mono")},
		{"alacritty", map[string]string{"ALACRITTY_WINDOW_ID": "1"}, present("Fira Code")},
		{"foot", map[string]string{"TERM": "foot"}, present("Hack")},
		{"wezterm in home", map[string]string{"WEZTERM_EXECUTABLE": "/usr/bin/wezterm"}, present("Cascadia Code")},
		{"ghostty", map[string]string{"GHOSTTY_RESOURCES_DIR": "/usr/share/ghostty"}, present("Berkeley Mono")},
		{"unsupported terminal", map[string]string{"TERM_PROGRAM": "vscode"}, Value[string]{}},
		{"no font set", map[string]string{"KONSOLE_VERSION": "1"}, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{"HOME": "/home/u"}
			for k, v := range tt.env {
				env[k] = v
			}
			src := Source{FS: files, Getenv: func(k string) string { return env[k] }}
			if got := Collect(context.Background(), src, []Fact{FactTerminalFont}, Options{}).TerminalFont; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
	if got := collectOne(t, sshServer, FactTerminalFont).TerminalFont; got != absent[string]() {
		t.Errorf("over SSH: got %+v, want absent", got)
	}
}

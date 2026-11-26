package sysinfo

import "testing"

func TestAppearance(t *testing.T) {
	gnome := ubuntuDesktop
	gnome.commands = map[string]string{
		"gsettings get org.gnome.desktop.interface gtk-theme":    "'Yaru-dark'\n",
		"gsettings get org.gnome.desktop.interface icon-theme":   "'Yaru'\n",
		"gsettings get org.gnome.desktop.interface font-name":    "'Ubuntu Sans 11'\n",
		"gsettings get org.gnome.desktop.interface cursor-theme": "''\n",
	}
	kde := btrfsBox
	kde.commands = map[string]string{
		"kreadconfig5 --file kdeglobals --group KDE --key widgetStyle": "Breeze\n",
		"kreadconfig6 --file kdeglobals --group General --key font":    "Noto Sans,10,-1,5,50,0,0,0,0,0\n",
	}
	tests := []struct {
		name    string
		machine machine
		fact    Fact
		want    Value[string]
	}{
		{"gsettings", gnome, FactTheme, present("Yaru-dark")},
		{"gsettings icons", gnome, FactIconTheme, present("Yaru")},
		{"gsettings font", gnome, FactFont, present("Ubuntu Sans 11")},
		{"empty gsettings falls back to gtk", gnome, FactCursor, present("Bibata-Modern-Classic")},
		{"kreadconfig5", kde, FactTheme, present("Breeze")},
		{"kreadconfig6 font", kde, FactFont, present("Noto Sans 10")},
		{"kde without a value", kde, FactCursor, Value[string]{}},
		{"gtk 3 settings", archLaptop, FactIconTheme, present("Papirus-Dark")},
		{"gtk 4 wins", archLaptop, FactTheme, present("Adwaita-dark")},
		{"no session", sshServer, FactTheme, absent[string]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := collectOne(t, tt.machine, tt.fact)
			if got := *appearanceField(&info, tt.fact); got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

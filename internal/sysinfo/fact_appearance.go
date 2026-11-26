package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strings"
)

// Facts about the look of the desktop.
const (
	// FactTheme is the GTK or widget theme.
	FactTheme Fact = "theme"
	// FactIconTheme is the icon theme.
	FactIconTheme Fact = "icon_theme"
	// FactFont is the interface font.
	FactFont Fact = "font"
	// FactCursor is the cursor theme.
	FactCursor Fact = "cursor"
)

// appearanceSetting tells where one setting is stored by each desktop.
type appearanceSetting struct {
	// gsettings is the key in org.gnome.desktop.interface.
	gsettings string
	// kdeFile, kdeGroup and kdeKey locate the KDE setting.
	kdeFile, kdeGroup, kdeKey string
	// gtk is the key in the settings.ini of GTK 3 and 4.
	gtk string
}

// appearanceSettings maps the appearance facts to their storage.
var appearanceSettings = map[Fact]appearanceSetting{
	FactTheme:     {"gtk-theme", "kdeglobals", "KDE", "widgetStyle", "gtk-theme-name"},
	FactIconTheme: {"icon-theme", "kdeglobals", "Icons", "Theme", "gtk-icon-theme-name"},
	FactFont:      {"font-name", "kdeglobals", "General", "font", "gtk-font-name"},
	FactCursor:    {"cursor-theme", "kcminputrc", "Mouse", "cursorTheme", "gtk-cursor-theme-name"},
}

// gnomeFamily are desktops that keep their settings in the GNOME schema.
var gnomeFamily = []string{"gnome", "unity", "pantheon", "budgie", "cinnamon"}

func init() {
	for f, setting := range appearanceSettings {
		get := func(ctx context.Context, s *session) Value[string] { return collectAppearance(ctx, s, setting) }
		collectors[f] = fact(get, func(i *Info) *Value[string] { return appearanceField(i, f) })
	}
}

// appearanceField returns the field of info that holds f.
func appearanceField(info *Info, f Fact) *Value[string] {
	switch f {
	case FactTheme:
		return &info.Theme
	case FactIconTheme:
		return &info.IconTheme
	case FactFont:
		return &info.Font
	default:
		return &info.Cursor
	}
}

// collectAppearance reads a setting from the desktop's own store: the
// GNOME schema through gsettings, the KDE configuration through
// kreadconfig, and otherwise the settings.ini of GTK 4 or 3. The fact is
// Absent without a graphical session.
func collectAppearance(ctx context.Context, s *session, setting appearanceSetting) Value[string] {
	if sessionType(s.getenv).State != Present {
		return absent[string]()
	}
	desktop := strings.ToLower(s.getenv("XDG_CURRENT_DESKTOP"))
	for _, name := range gnomeFamily {
		if !strings.Contains(desktop, name) {
			continue
		}
		out, err := s.run(ctx, "gsettings", "get", "org.gnome.desktop.interface", setting.gsettings)
		if value := strings.Trim(strings.TrimSpace(string(out)), "'"); err == nil && value != "" {
			return present(value)
		}
		break
	}
	if strings.Contains(desktop, "kde") {
		for _, tool := range []string{"kreadconfig6", "kreadconfig5"} {
			out, err := s.run(ctx, tool, "--file", setting.kdeFile, "--group", setting.kdeGroup, "--key", setting.kdeKey)
			if value := strings.TrimSpace(string(out)); err == nil && value != "" {
				if setting.kdeKey == "font" {
					value = kdeFont(value)
				}
				return present(value)
			}
		}
	}
	for _, version := range []string{"gtk-4.0", "gtk-3.0"} {
		if value := iniValue(s.fsys(), path.Join(s.configHome(), version, "settings.ini"), setting.gtk); value != "" {
			return present(value)
		}
	}
	return Value[string]{}
}

// kdeFont turns a Qt font description such as
// "Noto Sans,10,-1,5,50,0,0,0,0,0" into "Noto Sans 10".
func kdeFont(desc string) string {
	fields := strings.Split(desc, ",")
	if len(fields) < 2 {
		return desc
	}
	return strings.TrimSpace(fields[0]) + " " + strings.TrimSpace(fields[1])
}

// iniValue returns the value of key in the INI file name, ignoring
// sections, or "" when the file or the key is missing.
func iniValue(fsys fs.FS, name, key string) string {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		k, v, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(k) == key {
			return strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return ""
}

// configHome returns the user's configuration directory as a path for the
// Source's file system: XDG_CONFIG_HOME when it is absolute, otherwise
// .config in the home directory. It returns "" when neither is known.
func (s *session) configHome() string {
	if dir := s.getenv("XDG_CONFIG_HOME"); strings.HasPrefix(dir, "/") {
		return strings.TrimPrefix(path.Clean(dir), "/")
	}
	if home := s.home(); home != "" {
		return path.Join(home, ".config")
	}
	return ""
}

package sysinfo

import "strings"

// desktop detects the window manager and the display server. The window
// manager is taken from the first source that knows it: the root window
// property (xprop), the window manager itself (wmctrl), then the
// XDG_CURRENT_DESKTOP variable.
func desktop(run runner, getenv func(string) string) Desktop {
	result := Desktop{Server: X11}
	if getenv("WAYLAND_DISPLAY") != "" {
		result.Server = Wayland
	}

	if out, err := run("xprop", "-root", "_NET_WM_NAME"); err == nil {
		result.WM = parseXprop(string(out))
	}
	if result.WM == "" {
		if out, err := run("wmctrl", "-m"); err == nil {
			result.WM = parseWmctrl(string(out))
		}
	}
	if result.WM == "" {
		result.WM = strings.TrimSpace(getenv("XDG_CURRENT_DESKTOP"))
	}
	return result
}

// parseXprop extracts the value from xprop output of the form
// `_NET_WM_NAME(UTF8_STRING) = "name"`. It returns "" when the property
// has no value.
func parseXprop(out string) string {
	_, value, found := strings.Cut(out, "=")
	if !found {
		return ""
	}
	return strings.Trim(value, " \"\n")
}

// parseWmctrl extracts the window manager name from the "Name:" line of
// `wmctrl -m` output. It returns "" when there is no such line.
func parseWmctrl(out string) string {
	for line := range strings.SplitSeq(out, "\n") {
		if name, found := strings.CutPrefix(line, "Name:"); found {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

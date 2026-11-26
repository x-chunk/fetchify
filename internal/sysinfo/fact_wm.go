package sysinfo

import (
	"context"
	"strings"
)

// FactWM is the window manager or Wayland compositor.
const FactWM Fact = "wm"

func init() {
	collectors[FactWM] = fact(collectWM, func(i *Info) *Value[string] { return &i.WM })
}

// processDisplayName pairs a process name, as truncated by the kernel, with the
// name to show for it.
type processDisplayName struct{ process, name string }

// waylandCompositors are matched against the process table in this order.
var waylandCompositors = []processDisplayName{
	{"sway", "sway"},
	{"Hyprland", "Hyprland"},
	{"kwin_wayland", "KWin"},
	{"gnome-shell", "Mutter"},
	{"river", "river"},
	{"niri", "niri"},
	{"labwc", "labwc"},
	{"wayfire", "wayfire"},
	{"weston", "weston"},
	{"cosmic-comp", "COSMIC"},
}

// x11WindowManagers are matched against the process table in this order
// when the X server cannot be asked.
var x11WindowManagers = []processDisplayName{
	{"i3", "i3"},
	{"bspwm", "bspwm"},
	{"openbox", "Openbox"},
	{"awesome", "awesome"},
	{"xfwm4", "Xfwm4"},
	{"kwin_x11", "KWin"},
	{"dwm", "dwm"},
	{"qtile", "Qtile"},
	{"xmonad", "xmonad"},
	{"herbstluftwm", "herbstluftwm"},
}

// collectWM finds the window manager. On Wayland the compositor is found
// in the process table. On X11 the X server is asked through xprop or
// wmctrl, and the process table is the last resort. The fact is Absent
// without a graphical session and Unknown when nothing is found.
func collectWM(ctx context.Context, s *session) Value[string] {
	session := sessionType(s.getenv)
	if session.State != Present {
		return absent[string]()
	}
	if session.V == Wayland {
		return matchProcess(s.processNames(), waylandCompositors)
	}
	if name, ok := wmFromXprop(ctx, s); ok {
		return present(name)
	}
	if out, err := s.run(ctx, "wmctrl", "-m"); err == nil {
		if name := parseWmctrl(string(out)); name != "" {
			return present(name)
		}
	}
	return matchProcess(s.processNames(), x11WindowManagers)
}

// matchProcess returns the display name of the first candidate that is
// running. Names that the kernel truncated, such as "xmonad-x86_64-l",
// match by prefix.
func matchProcess(running map[string]bool, candidates []processDisplayName) Value[string] {
	for _, c := range candidates {
		if running[c.process] {
			return present(c.name)
		}
	}
	for _, c := range candidates {
		for name := range running {
			if strings.HasPrefix(name, c.process+"-") {
				return present(c.name)
			}
		}
	}
	return Value[string]{}
}

// wmFromXprop asks the X server for the window that an EWMH window
// manager publishes on the root window, and then for that window's name.
func wmFromXprop(ctx context.Context, s *session) (string, bool) {
	out, err := s.run(ctx, "xprop", "-root", "_NET_SUPPORTING_WM_CHECK")
	if err != nil {
		return "", false
	}
	id := parseXpropWindow(string(out))
	if id == "" {
		return "", false
	}
	out, err = s.run(ctx, "xprop", "-id", id, "_NET_WM_NAME")
	if err != nil {
		return "", false
	}
	name := parseXprop(string(out))
	return name, name != ""
}

// parseXpropWindow extracts the window ID from xprop output of the form
// "_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x1a00003". It returns
// "" when the property is not set.
func parseXpropWindow(out string) string {
	_, id, found := strings.Cut(out, "window id #")
	if !found {
		return ""
	}
	id = strings.TrimSpace(id)
	if !strings.HasPrefix(id, "0x") || strings.ContainsAny(id, " \n") {
		return ""
	}
	return id
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

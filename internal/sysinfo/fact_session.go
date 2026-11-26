package sysinfo

import (
	"context"
	"strings"
)

// FactSession is the protocol of the graphical session.
const FactSession Fact = "session"

// SessionType is the display server protocol of a graphical session.
type SessionType string

// The session types fetchify can tell apart.
const (
	X11     SessionType = "X11"
	Wayland SessionType = "Wayland"
)

func init() {
	collectors[FactSession] = fact(collectSession, func(i *Info) *Value[SessionType] { return &i.Session })
}

// collectSession returns the session type. The fact is Absent without a
// graphical session, for example on a console or over SSH without X
// forwarding.
func collectSession(_ context.Context, s *session) Value[SessionType] {
	return sessionType(s.getenv)
}

// sessionType detects the session from the environment. XDG_SESSION_TYPE
// is the most specific source; a session started from a console with
// startx reports "tty" there and is found through DISPLAY instead.
func sessionType(getenv func(string) string) Value[SessionType] {
	switch strings.ToLower(getenv("XDG_SESSION_TYPE")) {
	case "wayland":
		return present(Wayland)
	case "x11":
		return present(X11)
	}
	switch {
	case getenv("WAYLAND_DISPLAY") != "":
		return present(Wayland)
	case getenv("DISPLAY") != "":
		return present(X11)
	}
	return absent[SessionType]()
}

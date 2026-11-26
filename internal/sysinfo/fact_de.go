package sysinfo

import (
	"context"
	"strings"
)

// FactDE is the desktop environment.
const FactDE Fact = "de"

func init() {
	collectors[FactDE] = fact(collectDE, func(i *Info) *Value[string] { return &i.DE })
}

// collectDE reads the desktop environment from XDG_CURRENT_DESKTOP, which
// is a colon-separated list whose first entry is the most specific, and
// otherwise from DESKTOP_SESSION. The fact is Absent without a graphical
// session, and also when neither variable is set: a session that runs a
// bare window manager has no desktop environment.
func collectDE(_ context.Context, s *session) Value[string] {
	if sessionType(s.getenv).State != Present {
		return absent[string]()
	}
	name, _, _ := strings.Cut(s.getenv("XDG_CURRENT_DESKTOP"), ":")
	if name = strings.TrimSpace(name); name == "" {
		name = strings.TrimSpace(s.getenv("DESKTOP_SESSION"))
	}
	if name == "" {
		return absent[string]()
	}
	return present(name)
}

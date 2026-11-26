package sysinfo

import (
	"context"
	"strings"
	"time"
)

// Facts about the locale and the clock.
const (
	// FactLocale is the locale of messages.
	FactLocale Fact = "locale"
	// FactTimezone is the time zone of the system.
	FactTimezone Fact = "timezone"
	// FactDate is the current date and time.
	FactDate Fact = "date"
)

func init() {
	collectors[FactLocale] = fact(collectLocale, func(i *Info) *Value[string] { return &i.Locale })
	collectors[FactTimezone] = fact(collectTimezone, func(i *Info) *Value[string] { return &i.Timezone })
	collectors[FactDate] = fact(collectDate, func(i *Info) *Value[time.Time] { return &i.Date })
}

// collectLocale returns the first of LC_ALL, LC_MESSAGES and LANG that is
// set, which is the precedence the C library uses for messages. The fact
// is Absent when none is set: the program then runs in the C locale.
func collectLocale(_ context.Context, s *session) Value[string] {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := s.getenv(key); v != "" {
			return present(v)
		}
	}
	return absent[string]()
}

// collectTimezone returns the IANA name of the time zone, for example
// "Europe/Berlin". It is taken from TZ, then from the target of
// etc/localtime, then from etc/timezone, which Debian keeps.
func collectTimezone(_ context.Context, s *session) Value[string] {
	if tz := strings.TrimPrefix(s.getenv("TZ"), ":"); tz != "" && !strings.HasPrefix(tz, "/") {
		return present(tz)
	}
	if s.src.Readlink != nil {
		if target, err := s.src.Readlink("etc/localtime"); err == nil {
			if _, name, found := strings.Cut(target, "zoneinfo/"); found && name != "" {
				return present(name)
			}
		}
	}
	if name, err := readTrimmed(s.fsys(), "etc/timezone"); err == nil && name != "" {
		return present(name)
	}
	return Value[string]{}
}

// collectDate returns the current time of the Source.
func collectDate(_ context.Context, s *session) Value[time.Time] {
	if s.src.Now == nil {
		return Value[time.Time]{}
	}
	return present(s.src.Now())
}

package distro

import (
	"strings"
)

// Release is the parsed content of os-release, as described in
// os-release(5). Fields that the file does not set are empty.
type Release struct {
	// ID identifies the distribution in lower case, for example "ubuntu".
	ID string
	// IDLike lists the identifiers of related distributions in lower
	// case, closest first, for example ["debian"].
	IDLike []string
	// Name is the name of the distribution without a version, for
	// example "Ubuntu".
	Name string
	// PrettyName is the name meant for display, usually with the
	// version, for example "Ubuntu 24.04 LTS".
	PrettyName string
	// VersionID is the version in a form meant for parsing, for example
	// "24.04".
	VersionID string
	// BuildID identifies the build of a rolling release.
	BuildID string
	// VariantID identifies the variant of the distribution, for example
	// "workstation".
	VariantID string
}

// ParseRelease parses the content of an os-release file. It never fails:
// unknown keys and malformed lines are skipped, and for a key that occurs
// several times the last value wins.
//
// Values may be unquoted or enclosed in single or double quotes. Inside
// double quotes and in unquoted values a backslash escapes the next
// character. Lines that start with '#' are comments.
func ParseRelease(data []byte) Release {
	var r Release
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		key, raw, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		value, ok := unquote(strings.TrimSpace(raw))
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "ID":
			r.ID = strings.ToLower(value)
		case "ID_LIKE":
			r.IDLike = strings.Fields(strings.ToLower(value))
		case "NAME":
			r.Name = value
		case "PRETTY_NAME":
			r.PrettyName = value
		case "VERSION_ID":
			r.VersionID = value
		case "BUILD_ID":
			r.BuildID = value
		case "VARIANT_ID":
			r.VariantID = value
		}
	}
	return r
}

// unquote returns the value of a shell-style assignment. It reports false
// for a quoted value whose closing quote is missing. Text after the
// closing quote is ignored.
func unquote(s string) (string, bool) {
	if s == "" {
		return "", true
	}
	switch s[0] {
	case '\'':
		end := strings.IndexByte(s[1:], '\'')
		if end < 0 {
			return "", false
		}
		return s[1 : 1+end], true
	case '"':
		var b strings.Builder
		for i := 1; i < len(s); i++ {
			switch c := s[i]; {
			case c == '"':
				return b.String(), true
			case c == '\\' && i+1 < len(s):
				i++
				b.WriteByte(s[i])
			default:
				b.WriteByte(c)
			}
		}
		return "", false
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String(), true
}

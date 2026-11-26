package sysinfo

import (
	"context"
	"io/fs"
)

// FactOS is the content of os-release, which identifies the
// distribution. sysinfo does not interpret it; distro.ParseRelease does.
const FactOS Fact = "os"

// osReleasePaths are the locations of os-release in order of precedence,
// as os-release(5) defines them.
var osReleasePaths = []string{"etc/os-release", "usr/lib/os-release"}

func init() {
	collectors[FactOS] = fact(collectOS, func(i *Info) *Value[[]byte] { return &i.OS })
}

// collectOS reads the first os-release file that exists. The fact is
// Unknown when there is none.
func collectOS(_ context.Context, s *session) Value[[]byte] {
	for _, name := range osReleasePaths {
		if data, err := fs.ReadFile(s.fsys(), name); err == nil {
			return present(data)
		}
	}
	return Value[[]byte]{}
}

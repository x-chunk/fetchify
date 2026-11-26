package sysinfo

import (
	"context"
	"path"
)

// FactShell is the login shell of the user.
const FactShell Fact = "shell"

func init() {
	collectors[FactShell] = fact(collectShell, func(i *Info) *Value[string] { return &i.Shell })
}

// collectShell returns the base name of the shell named by $SHELL. The
// fact is Unknown when the variable is unset.
func collectShell(_ context.Context, s *session) Value[string] {
	value := s.getenv("SHELL")
	if value == "" {
		return Value[string]{}
	}
	return present(path.Base(value))
}

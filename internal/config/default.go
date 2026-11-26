package config

import (
	_ "embed"
	"slices"
)

//go:embed default.toml
var defaultTOML []byte

// DefaultTOML returns the documented default configuration file. It is
// what --print-config prints, and it decodes to exactly Defaults().
func DefaultTOML() []byte {
	return slices.Clone(defaultTOML)
}

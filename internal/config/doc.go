// Package config holds the configuration of fetchify: its type, the
// defaults, the documented default file and the loading of TOML files
// with strict validation.
//
// The package knows neither the items of the screen nor the logos;
// callers pass the known identifiers in, so that config does not depend
// on the presentation layer.
package config

package sysinfo

import (
	"context"
	"strings"
	"unicode"
)

// FactHost is the model of the machine.
const FactHost Fact = "host"

func init() {
	collectors[FactHost] = fact(collectHost, func(i *Info) *Value[string] { return &i.Host })
}

// dmiPlaceholders are values firmware vendors leave in DMI fields they
// did not fill in. They are compared in lower case.
var dmiPlaceholders = map[string]bool{
	"":                       true,
	"to be filled by o.e.m.": true,
	"system product name":    true,
	"system version":         true,
	"not specified":          true,
	"default string":         true,
	"not applicable":         true,
	"none":                   true,
	"unknown":                true,
}

// isPlaceholder reports whether a DMI value carries no information.
func isPlaceholder(s string) bool {
	return dmiPlaceholders[strings.ToLower(strings.TrimSpace(s))]
}

// collectHost reads the model from DMI and, on machines without DMI such
// as ARM boards, from the device tree. The fact is Unknown when every
// source is missing or a placeholder.
func collectHost(_ context.Context, s *session) Value[string] {
	name, _ := readTrimmed(s.fsys(), "sys/devices/virtual/dmi/id/product_name")
	version, _ := readTrimmed(s.fsys(), "sys/devices/virtual/dmi/id/product_version")
	if model := hostModel(name, version); model != "" {
		return present(model)
	}
	model, err := readTrimmed(s.fsys(), "sys/firmware/devicetree/base/model")
	if model = strings.TrimRight(model, "\x00"); err == nil && !isPlaceholder(model) {
		return present(model)
	}
	return Value[string]{}
}

// hostModel combines the DMI product name and version into one model
// string, or returns "" when both are placeholders.
//
// The version is appended when it adds something. Lenovo stores a model
// code such as "21HMCTO1WW" in the name and the readable name in the
// version; then the version alone is the model.
func hostModel(name, version string) string {
	if isPlaceholder(name) {
		name = ""
	}
	if isPlaceholder(version) {
		version = ""
	}
	switch {
	case name == "":
		return version
	case version == "" || strings.Contains(name, version):
		return name
	case isModelCode(name) && strings.Contains(version, " "):
		return version
	}
	return name + " " + version
}

// isModelCode reports whether s looks like a machine type code rather
// than a name: one word of upper-case letters and digits with both.
func isModelCode(s string) bool {
	var letters, digits bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			letters = true
		case unicode.IsDigit(r):
			digits = true
		default:
			return false
		}
	}
	return letters && digits && len(s) >= 6
}

// Package sysinfo collects facts about the machine fetchify runs on.
//
// It returns plain data only: no formatting, no colors and no fallback
// strings. A fact that cannot be determined is left at its zero value, and
// deciding how to show that is the job of the presentation layer.
package sysinfo

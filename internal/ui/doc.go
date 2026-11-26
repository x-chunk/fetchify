// Package ui is the presentation layer of fetchify: it turns a
// sysinfo.Info snapshot into a beautify.Template.
//
// The package decides what is shown, in which order, with which wording
// and which styles. It collects nothing and writes nothing, which keeps it
// a pure function of its input.
//
// The screen is a row of the logo and an info column. The column stacks a
// header, a key-value list and a palette.
package ui

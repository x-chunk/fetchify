// Package ui is the presentation layer of fetchify: it turns a
// sysinfo.Info snapshot into a render.Block.
//
// The package decides what is shown, in which order, with which wording
// and which styles. It collects nothing and writes nothing, which keeps it
// a pure function of its input.
//
// The screen is assembled from small parts that correspond to beautify
// blocks, to keep the later migration mechanical: the logo and the info
// column form a row, the info column stacks a header, a key-value list
// and a palette.
package ui

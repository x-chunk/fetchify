// Package ui is the presentation layer of fetchify: it turns the
// collected facts and the configuration into the screen, a
// beautify.Composable.
//
// The package decides what is shown, in which order, with which wording
// and which styles. It collects, measures and writes nothing, which keeps
// it a pure function of its Input.
//
// The screen is the logo of the distribution next to a card. The card
// has one section per configured group, and every item of a group adds
// rows to it. Items are registered in a table keyed by the identifiers
// the configuration uses; Facts tells which facts a configuration needs.
package ui

// Package render is the low-level output layer of fetchify: styled text,
// the operations that lay it out, and the Engine that encodes it as bytes.
//
// It is a deliberately small stand-in for github.com/x-chunk/beautify/dra
// and follows the same rules, so that the migration is a mechanical swap:
//
//   - Span, Line and Block are immutable values. Operations return new
//     values and never modify their arguments.
//   - Text is sanitized when a Span is created, so external input cannot
//     inject escape sequences.
//   - Width is measured as one cell per rune (Real Mode in beautify terms).
//   - The Engine is pure and stateless. It never inspects the terminal or
//     the environment; the caller chooses the Profile.
//
// Nothing outside this package produces escape sequences.
package render

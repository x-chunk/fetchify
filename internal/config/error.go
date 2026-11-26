package config

import (
	"fmt"
	"strings"
)

// Problem is one thing wrong with a configuration file.
type Problem struct {
	// Line and Column locate the problem, from 1. They are 0 when the
	// position is not known; the message then names the full key.
	Line, Column int
	// Message describes the problem and names the key it concerns.
	Message string
}

// String formats p as "line:column: message", or as the message alone
// when the position is not known.
func (p Problem) String() string {
	if p.Line == 0 {
		return p.Message
	}
	return fmt.Sprintf("%d:%d: %s", p.Line, p.Column, p.Message)
}

// Error lists every problem found in a configuration file.
type Error struct {
	// Path is the file the problems were found in.
	Path string
	// Problems are sorted by position; problems without one come last.
	Problems []Problem
}

// Error returns one line per problem in the form
// "path:line:column: message".
func (e *Error) Error() string {
	lines := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		lines[i] = e.Path + ":" + p.String()
	}
	return strings.Join(lines, "\n")
}

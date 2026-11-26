// Package cli parses the command line of fetchify and holds its help
// text.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Options are the parsed flags.
type Options struct {
	// Help asks for the usage text.
	Help bool
	// DontClear keeps the screen instead of clearing it.
	DontClear bool
}

// ErrUsage wraps every error of Parse: the command line is invalid and
// the caller should print the usage.
var ErrUsage = errors.New("invalid command line")

// Parse parses the arguments without the program name. It never prints;
// the error describes the problem and wraps ErrUsage.
func Parse(args []string) (Options, error) {
	var o Options
	fs := flag.NewFlagSet("fetchify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolVar(&o.Help, "help", false, "")
	fs.BoolVar(&o.DontClear, "dont-clear", false, "")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{Help: true}, nil
		}
		return Options{}, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	if fs.NArg() > 0 {
		return Options{}, fmt.Errorf("%w: unexpected argument %q", ErrUsage, fs.Arg(0))
	}
	return o, nil
}

// Usage returns the help text.
func Usage() string {
	return strings.TrimLeft(`
Usage: fetchify [flags]

Prints facts about this machine next to the logo of its distribution.

Flags:
  --dont-clear  do not clear the screen before printing
  --help        show this help
`, "\n")
}

// Package cli parses the command line of fetchify and holds its help
// text.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"go.xchunk.org/fetchify/internal/config"
)

// Options are the parsed flags.
type Options struct {
	// Help asks for the usage text.
	Help bool
	// Clear clears the screen before printing.
	Clear bool
	// ConfigPath is the configuration file to read instead of the
	// default one, or "".
	ConfigPath string
	// Logo overrides logo.name, or is "".
	Logo string
	// NoIcons turns the icons off.
	NoIcons bool
	// NoColor turns colors off.
	NoColor bool
	// PrintConfig asks for the documented default configuration.
	PrintConfig bool
}

// Apply writes the overrides of the flags into cfg. Flags win over the
// configuration file.
func (o Options) Apply(cfg *config.Config) {
	if o.Logo != "" {
		cfg.Logo.Name = o.Logo
	}
	if o.NoIcons {
		cfg.Icons.Enabled = false
	}
	if o.NoColor {
		cfg.Colors.Mode = config.ModeNone
	}
	if o.Clear {
		cfg.Clear = true
	}
}

// ErrUsage wraps every error of Parse: the command line is invalid and
// the caller should print the usage.
var ErrUsage = errors.New("invalid command line")

// Parse parses the arguments without the program name. logos are the
// names --logo accepts besides "auto" and "none". Parse never prints; the
// error describes the problem and wraps ErrUsage.
func Parse(args []string, logos []string) (Options, error) {
	var o Options
	fs := flag.NewFlagSet("fetchify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolVar(&o.Help, "help", false, "")
	fs.BoolVar(&o.Clear, "clear", false, "")
	// --dont-clear is accepted and ignored so that shell rc files written
	// when clearing was the default keep working. It is not in the help.
	fs.Bool("dont-clear", false, "")
	fs.StringVar(&o.ConfigPath, "config", "", "")
	fs.StringVar(&o.Logo, "logo", "", "")
	fs.BoolVar(&o.NoIcons, "no-icons", false, "")
	fs.BoolVar(&o.NoColor, "no-color", false, "")
	fs.BoolVar(&o.PrintConfig, "print-config", false, "")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{Help: true}, nil
		}
		return Options{}, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	if fs.NArg() > 0 {
		return Options{}, fmt.Errorf("%w: unexpected argument %q", ErrUsage, fs.Arg(0))
	}
	if o.Logo != "" && o.Logo != config.LogoAuto && o.Logo != config.LogoNone && !slices.Contains(logos, o.Logo) {
		return Options{}, fmt.Errorf("%w: unknown logo %q; known logos: auto, none, %s", ErrUsage, o.Logo, strings.Join(logos, ", "))
	}
	return o, nil
}

// Usage returns the help text.
func Usage() string {
	return strings.TrimLeft(`
Usage: fetchify [flags]

Prints facts about this machine next to the logo of its distribution.

Flags:
  --config <path>  read this configuration file instead of
                   $XDG_CONFIG_HOME/fetchify/config.toml
  --logo <name>    show this logo: auto, none, koala or a distribution
  --no-icons       do not show Nerd Font icons
  --no-color       do not use colors; NO_COLOR has the same effect
  --clear          clear the screen before printing
  --print-config   print the documented default configuration and exit
  --help           show this help and exit
`, "\n")
}

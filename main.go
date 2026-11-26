// Command fetchify prints a summary of the machine it runs on next to an
// ASCII-art logo.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/cli"
	"go.xchunk.org/fetchify/internal/config"
	"go.xchunk.org/fetchify/internal/distro"
	"go.xchunk.org/fetchify/internal/sysinfo"
	"go.xchunk.org/fetchify/internal/ui"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes fetchify with the command-line arguments args, writing the
// screen to stdout and diagnostics to stderr. It returns the process exit
// code: 0 on success, 1 when the output cannot be written and 2 for
// invalid arguments.
//
// This function is the only place that writes program output; the layers
// below it only build values.
func run(args []string, stdout, stderr io.Writer) int {
	opts, err := cli.Parse(args, distro.IDs())
	if err != nil {
		fmt.Fprintln(stderr, "fetchify:", err)
		fmt.Fprint(stderr, cli.Usage())
		return 2
	}
	// Help and the default configuration do not read the configuration
	// file, so they work when it is broken.
	if opts.Help {
		fmt.Fprint(stdout, cli.Usage())
		return 0
	}
	if opts.PrintConfig {
		if _, err := stdout.Write(config.DefaultTOML()); err != nil {
			fmt.Fprintln(stderr, "fetchify:", err)
			return 1
		}
		return 0
	}

	cfg, err := loadConfig(opts, os.Getenv)
	if err != nil {
		var problems *config.Error
		if errors.As(err, &problems) {
			for _, line := range strings.Split(problems.Error(), "\n") {
				fmt.Fprintln(stderr, "fetchify:", line)
			}
		} else {
			fmt.Fprintln(stderr, "fetchify:", err)
		}
		return 2
	}
	opts.Apply(&cfg)

	start := time.Now()
	info := sysinfo.Collect(context.Background(), sysinfo.SystemSource(), ui.Facts(cfg), ui.Options(cfg))
	elapsed := time.Since(start)

	if cfg.Clear && isTerminal(stdout) {
		// Home, erase the screen and erase the scrollback. This is the one
		// escape sequence fetchify writes itself.
		if _, err := io.WriteString(stdout, "\x1b[H\x1b[2J\x1b[3J"); err != nil {
			fmt.Fprintln(stderr, "fetchify:", err)
			return 1
		}
	}
	composer := beautify.NewComposer(stdout, composerOptions(cfg.Colors.Mode)...)
	screen := beautify.NewTemplate(ui.Screen(ui.Input{Info: info, Config: cfg, Elapsed: elapsed}))
	if err := composer.Render(screen); err != nil {
		fmt.Fprintln(stderr, "fetchify:", err)
		return 1
	}
	return 0
}

// profiles maps the fixed color modes to profiles. "none" is plain text:
// with colors off, attributes such as reverse video would be the only
// escape sequences left, and they are not worth them.
var profiles = map[config.ColorMode]dra.Profile{
	config.ModeTrueColor: dra.ProfileTrueColor,
	config.Mode256:       dra.ProfileANSI256,
	config.Mode16:        dra.ProfileANSI16,
	config.ModeNone:      dra.ProfilePlain,
}

// composerOptions returns the options of the Composer for a color mode.
// For "auto" the Composer detects the profile itself, which honors
// NO_COLOR and writes plain text when the output is not a terminal.
func composerOptions(mode config.ColorMode) []beautify.ComposerOption {
	if p, ok := profiles[mode]; ok {
		return []beautify.ComposerOption{beautify.WithProfile(p)}
	}
	return nil
}

// loadConfig reads the configuration file named by --config, or the
// default one. A default file that does not exist is normal and gives
// the defaults; a file named on the command line must exist.
func loadConfig(opts cli.Options, getenv func(string) string) (config.Config, error) {
	known := config.Known{Items: ui.ItemIDs(), Logos: distro.IDs()}
	if opts.ConfigPath != "" {
		return config.Load(opts.ConfigPath, known)
	}
	home, _ := os.UserHomeDir()
	path := config.DefaultPath(getenv, home)
	if path == "" {
		return config.Defaults(), nil
	}
	cfg, err := config.Load(path, known)
	if errors.Is(err, fs.ErrNotExist) {
		return config.Defaults(), nil
	}
	return cfg, err
}

// isTerminal reports whether w is a terminal. The screen is never
// cleared when the output goes elsewhere.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

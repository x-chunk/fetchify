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
	"os/exec"
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

	if !opts.DontClear {
		clearScreen(stdout)
	}

	start := time.Now()
	info := sysinfo.Collect(context.Background(), sysinfo.SystemSource(), ui.Facts(cfg), sysinfo.Options{})
	elapsed := time.Since(start)

	// The profile is fixed until it is detected from the terminal: ANSI16
	// is what fetchify has always written.
	composer := beautify.NewComposer(stdout, beautify.WithProfile(dra.ProfileANSI16))
	screen := beautify.NewTemplate(ui.Screen(ui.Input{Info: info, Config: cfg, Elapsed: elapsed}))
	if err := composer.Render(screen); err != nil {
		fmt.Fprintln(stderr, "fetchify:", err)
		return 1
	}
	return 0
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

// clearScreen clears the terminal behind w with the clear utility. A
// failure is ignored: the screen is then printed below the existing
// content.
func clearScreen(w io.Writer) {
	cmd := exec.Command("clear")
	cmd.Stdout = w
	_ = cmd.Run()
}

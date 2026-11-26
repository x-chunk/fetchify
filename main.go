// Command fetchify prints a summary of the machine it runs on next to an
// ASCII-art logo.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"

	"go.xchunk.org/fetchify/internal/cli"
	"go.xchunk.org/fetchify/internal/config"
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
	opts, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, "fetchify:", err)
		fmt.Fprint(stderr, cli.Usage())
		return 2
	}
	if opts.Help {
		fmt.Fprint(stdout, cli.Usage())
		return 0
	}

	if !opts.DontClear {
		clearScreen(stdout)
	}

	cfg := config.Defaults()
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

// clearScreen clears the terminal behind w with the clear utility. A
// failure is ignored: the screen is then printed below the existing
// content.
func clearScreen(w io.Writer) {
	cmd := exec.Command("clear")
	cmd.Stdout = w
	_ = cmd.Run()
}

// Command fetchify prints a summary of the machine it runs on next to an
// ASCII-art logo.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/x-chunk/fetchify/internal/render"
	"github.com/x-chunk/fetchify/internal/sysinfo"
	"github.com/x-chunk/fetchify/internal/ui"
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
	flags := flag.NewFlagSet("fetchify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	help := flags.Bool("help", false, "Show help message")
	dontClear := flags.Bool("dont-clear", false, "Don't clear screen before print fetch")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *help {
		fmt.Fprintln(stdout, "Use:")
		fmt.Fprintln(stdout, "  --help   Show help message")
		fmt.Fprintln(stdout, "  --dont-clear  Don't clear screen before print fetch")
		return 0
	}

	if !*dontClear {
		clearScreen(stdout)
	}

	screen := ui.Build(sysinfo.Collect())
	out := render.NewEngine(render.ProfileANSI16).Append(nil, screen)
	out = append(out, '\n')
	if _, err := stdout.Write(out); err != nil {
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

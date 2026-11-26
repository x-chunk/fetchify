package sysinfo

import (
	"context"
	"strings"
)

// FactTerminal is the terminal emulator the process runs in.
const FactTerminal Fact = "terminal"

func init() {
	collectors[FactTerminal] = fact(collectTerminal, func(i *Info) *Value[string] { return &i.Terminal })
}

// terminalVars maps variables that terminals set for their children to
// the name of the terminal, in the order they are checked.
var terminalVars = []struct{ key, name string }{
	{"KITTY_WINDOW_ID", "kitty"},
	{"ALACRITTY_WINDOW_ID", "Alacritty"},
	{"WEZTERM_EXECUTABLE", "WezTerm"},
	{"KONSOLE_VERSION", "Konsole"},
	{"GNOME_TERMINAL_SCREEN", "GNOME Terminal"},
	{"TILIX_ID", "Tilix"},
	{"GHOSTTY_RESOURCES_DIR", "Ghostty"},
}

// terminalWrappers are processes between a terminal and fetchify that
// are not the terminal: shells, privilege tools and login.
var terminalWrappers = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "fish": true, "dash": true, "nu": true,
	"ksh": true, "tcsh": true, "sudo": true, "doas": true, "su": true, "login": true,
}

// terminalNames maps process names, as truncated by the kernel, to
// display names. Other processes are reported by their name.
var terminalNames = map[string]string{
	"gnome-terminal-": "GNOME Terminal",
	"alacritty":       "Alacritty",
	"wezterm-gui":     "WezTerm",
	"konsole":         "Konsole",
	"tilix":           "Tilix",
	"ghostty":         "Ghostty",
	"xfce4-terminal":  "Xfce Terminal",
	"tmux: server":    "tmux",
	"tmux: client":    "tmux",
	"SCREEN":          "screen",
}

// maxTerminalSteps bounds the walk up the process tree.
const maxTerminalSteps = 16

// collectTerminal finds the terminal from the variables it sets and
// otherwise from the process tree. Multiplexers such as tmux are
// reported as the terminal: the real one is not reachable from inside.
// The fact is Absent on a Linux console, over SSH and without a
// terminal.
func collectTerminal(_ context.Context, s *session) Value[string] {
	if name := s.getenv("TERM_PROGRAM"); name != "" {
		return present(strings.TrimSuffix(name, ".app"))
	}
	for _, v := range terminalVars {
		if s.getenv(v.key) != "" {
			return present(v.name)
		}
	}
	if term := s.getenv("TERM"); term == "foot" || strings.HasPrefix(term, "foot-") {
		return present("foot")
	}
	return terminalFromProcesses(s)
}

// terminalFromProcesses walks up from the parent of the process and
// returns the first process that is not a wrapper. Reaching PID 1 or
// sshd means there is no terminal emulator.
func terminalFromProcesses(s *session) Value[string] {
	pid := s.src.PID
	for range maxTerminalSteps {
		ppid, err := parentPID(s.fsys(), pid)
		if err != nil {
			return Value[string]{}
		}
		if ppid <= 1 {
			return absent[string]()
		}
		pid = ppid
		name, err := processName(s.fsys(), pid)
		if err != nil {
			return Value[string]{}
		}
		switch {
		case terminalWrappers[name]:
			continue
		case name == "sshd" || strings.HasPrefix(name, "sshd-"):
			return absent[string]()
		case terminalNames[name] != "":
			return present(terminalNames[name])
		default:
			return present(name)
		}
	}
	return Value[string]{}
}

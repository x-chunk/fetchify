package sysinfo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// machine describes a fixture machine: a file tree under testdata and
// everything else a Source provides.
type machine struct {
	// dir is the directory under testdata that holds the file tree. An
	// empty dir gives a machine without files.
	dir string
	// env holds the environment variables.
	env map[string]string
	// commands maps a command line, words joined by single spaces, to its
	// output. Commands not listed fail as if they were not installed.
	commands map[string]string
	// uid and pid are the identity of the process.
	uid, pid int
	// users maps user IDs to login names.
	users map[int]string
	// mounts maps mount points to their sizes. Mount points not listed
	// fail to stat.
	mounts map[string]FSStat
}

// errNotFound is returned for commands and lookups a machine does not
// have.
var errNotFound = errors.New("not found")

// source returns the Source of m.
func (m machine) source() Source {
	src := Source{
		FS:     emptyFS{},
		Getenv: func(key string) string { return m.env[key] },
		UID:    m.uid,
		PID:    m.pid,
		Run: func(_ context.Context, name string, args ...string) ([]byte, error) {
			out, ok := m.commands[strings.Join(append([]string{name}, args...), " ")]
			if !ok {
				return nil, errNotFound
			}
			return []byte(out), nil
		},
		LookupUser: func(uid int) (string, error) {
			if name, ok := m.users[uid]; ok {
				return name, nil
			}
			return "", errNotFound
		},
		Statfs: func(path string) (FSStat, error) {
			if st, ok := m.mounts[path]; ok {
				return st, nil
			}
			return FSStat{}, errNotFound
		},
	}
	if m.dir != "" {
		src.FS = os.DirFS(filepath.Join("testdata", m.dir))
	}
	return src
}

// collectOne collects the single fact f from m.
func collectOne(t *testing.T, m machine, f Fact) Info {
	t.Helper()
	return Collect(context.Background(), m.source(), []Fact{f}, Options{})
}

// The fixture machines of the plan. Each port of a fact adds the files
// that fact reads to the trees under testdata.
var (
	// ubuntuDesktop is a typical Ubuntu desktop with GNOME on Wayland.
	ubuntuDesktop = machine{
		dir:   "ubuntu-desktop",
		uid:   1000,
		pid:   4242,
		users: map[int]string{1000: "alice"},
		env: map[string]string{
			"USER":  "alice",
			"HOME":  "/home/alice",
			"SHELL": "/usr/bin/zsh",
			"TERM":  "xterm-256color",

			"XDG_SESSION_TYPE":    "wayland",
			"WAYLAND_DISPLAY":     "wayland-0",
			"XDG_CURRENT_DESKTOP": "ubuntu:GNOME",

			"GNOME_TERMINAL_SCREEN": "/org/gnome/Terminal/screen/1",
		},
	}
	// archLaptop is an Arch laptop with a battery and two GPUs.
	archLaptop = machine{
		dir:   "arch-laptop",
		uid:   1000,
		pid:   777,
		users: map[int]string{1000: "bob"},
		env: map[string]string{
			"USER":    "bob",
			"SHELL":   "/bin/bash",
			"DISPLAY": ":0",
		},
		commands: map[string]string{
			"xprop -root _NET_SUPPORTING_WM_CHECK": "_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x1a00003\n",
			"xprop -id 0x1a00003 _NET_WM_NAME":     "_NET_WM_NAME(UTF8_STRING) = \"i3\"\n",
		},
	}
	// sshServer is a headless Debian server reached over SSH, logged in
	// as root.
	sshServer = machine{
		dir:   "ssh-server",
		uid:   0,
		pid:   900,
		users: map[int]string{0: "root"},
		env: map[string]string{
			"USER":    "root",
			"SHELL":   "/bin/bash",
			"SSH_TTY": "/dev/pts/0",
		},
	}
	// container is an Alpine container with an arbitrary UID that has no
	// entry in the user database: no DMI, no DRM, an overlay root.
	container = machine{
		dir: "container",
		uid: 100123,
		pid: 1,
	}
	// raspberryPi is a Raspberry Pi with a device tree and no DMI.
	raspberryPi = machine{
		dir:   "raspberry-pi",
		uid:   1000,
		pid:   1500,
		users: map[int]string{1000: "pi"},
		env:   map[string]string{"SHELL": "/bin/bash"},
	}
	// btrfsBox is a Fedora machine with btrfs subvolumes and bind mounts.
	btrfsBox = machine{
		dir:   "btrfs",
		uid:   1000,
		pid:   2000,
		users: map[int]string{1000: "carol"},
		env: map[string]string{
			"SHELL":               "/usr/bin/fish",
			"TERM_PROGRAM":        "tmux",
			"XDG_SESSION_TYPE":    "wayland",
			"XDG_CURRENT_DESKTOP": "KDE",
		},
		commands: map[string]string{
			"rpm -qa": "bash-5.2.32-1.fc41.x86_64\nkernel-6.11.4-301.fc41.x86_64\nfish-3.7.1-3.fc41.x86_64\n",
		},
	}
	// placeholderDMI is a self-built desktop whose firmware reports
	// placeholder strings everywhere.
	placeholderDMI = machine{
		dir:   "placeholder-dmi",
		uid:   1000,
		pid:   3000,
		users: map[int]string{1000: "dave"},
	}
)

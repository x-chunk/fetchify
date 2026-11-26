package sysinfo

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"time"
)

// Source is everything a collector may read. Collectors never touch the
// machine directly, so tests replace a Source with fixtures. A collector
// that needs something a Source cannot give gets a new field here.
//
// Every function field must be safe for concurrent use: collectors run in
// parallel. A nil function field behaves as if the facility were missing.
type Source struct {
	// FS is the file system rooted at "/". Paths are written without the
	// leading slash, as io/fs requires.
	FS fs.FS
	// Run runs the command name with args and returns its standard
	// output. It must stop the command when ctx is done.
	Run func(ctx context.Context, name string, args ...string) ([]byte, error)
	// Getenv returns the value of an environment variable, or "".
	Getenv func(key string) string
	// Statfs returns the sizes of the file system mounted at path.
	Statfs func(path string) (FSStat, error)
	// Now returns the current time.
	Now func() time.Time
	// UID is the effective user ID of the process.
	UID int
	// PID is the ID of the process.
	PID int
	// LookupUser returns the login name of the user with the ID uid.
	LookupUser func(uid int) (string, error)
	// Interfaces returns the network interfaces with their addresses.
	Interfaces func() ([]NetInterface, error)
	// Fetch sends an HTTP GET request for url and returns at most limit
	// bytes of the body of a successful response. It must stop when ctx
	// is done.
	Fetch func(ctx context.Context, url string, limit int64) ([]byte, error)
}

// NetInterface is a network interface and its addresses.
type NetInterface struct {
	// Name is the name of the interface, for example "wlan0".
	Name string
	// Up reports whether the interface is administratively up.
	Up bool
	// Loopback reports whether the interface is a loopback interface.
	Loopback bool
	// Addrs lists the addresses of the interface in kernel order.
	Addrs []netip.Addr
}

// FSStat holds the sizes of a mounted file system, in bytes.
type FSStat struct {
	// Total is the size of the file system.
	Total uint64
	// Free is the space not in use, including blocks reserved for root.
	Free uint64
	// Available is the space an unprivileged user can still use.
	Available uint64
}

// commandWaitDelay bounds how long a command may keep its output open
// after it was killed, for example through a child that inherited it.
const commandWaitDelay = 50 * time.Millisecond

// SystemSource returns the Source of the running machine.
func SystemSource() Source {
	return Source{
		FS:     os.DirFS("/"),
		Run:    runCommand,
		Getenv: os.Getenv,
		Statfs: statfs,
		Now:    time.Now,
		UID:    os.Geteuid(),
		PID:    os.Getpid(),
		LookupUser: func(uid int) (string, error) {
			u, err := user.LookupId(strconv.Itoa(uid))
			if err != nil {
				return "", err
			}
			return u.Username, nil
		},
		Interfaces: interfaces,
		Fetch:      fetch,
	}
}

// interfaces lists the network interfaces of the machine.
func interfaces() ([]NetInterface, error) {
	list, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := make([]NetInterface, 0, len(list))
	for _, iface := range list {
		ni := NetInterface{
			Name:     iface.Name,
			Up:       iface.Flags&net.FlagUp != 0,
			Loopback: iface.Flags&net.FlagLoopback != 0,
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if prefix, err := netip.ParsePrefix(a.String()); err == nil {
				ni.Addrs = append(ni.Addrs, prefix.Addr())
			}
		}
		result = append(result, ni)
	}
	return result, nil
}

// fetch sends a GET request with a generic User-Agent and nothing else
// that identifies the user, and reads at most limit bytes of the body.
func fetch(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "fetchify")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sysinfo: %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// runCommand runs name, resolved through PATH and never through a shell,
// and returns its standard output. Standard error is discarded. The
// command is killed when ctx is done.
func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = commandWaitDelay
	return cmd.Output()
}

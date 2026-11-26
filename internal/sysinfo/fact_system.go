package sysinfo

import (
	"bytes"
	"context"
	"encoding/binary"
	"io/fs"
	"strconv"
	"strings"
)

// Facts about the running system.
const (
	// FactInit is the name of the init process.
	FactInit Fact = "init"
	// FactProcesses is the number of processes.
	FactProcesses Fact = "processes"
	// FactUsers is the number of distinct users with a login session.
	FactUsers Fact = "users"
)

// The layout of a utmp record of glibc on 64-bit Linux, which is the
// same on 32-bit systems with 64-bit time.
const (
	utmpSize        = 384
	utmpUserOffset  = 44
	utmpUserSize    = 32
	utmpUserProcess = 7 // ut_type of a login session
)

func init() {
	collectors[FactInit] = fact(collectInit, func(i *Info) *Value[string] { return &i.Init })
	collectors[FactProcesses] = fact(collectProcesses, func(i *Info) *Value[int] { return &i.Processes })
	collectors[FactUsers] = fact(collectUsers, func(i *Info) *Value[int] { return &i.Users })
}

// collectInit reads the command name of PID 1, for example "systemd". In
// a container this is the entry point of the container.
func collectInit(_ context.Context, s *session) Value[string] {
	name, err := processName(s.fsys(), 1)
	if err != nil || name == "" {
		return Value[string]{}
	}
	return present(name)
}

// collectProcesses counts the process directories in proc.
func collectProcesses(_ context.Context, s *session) Value[int] {
	entries, err := fs.ReadDir(s.fsys(), "proc")
	if err != nil {
		return Value[int]{}
	}
	n := 0
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err == nil {
			n++
		}
	}
	if n == 0 {
		return Value[int]{}
	}
	return present(n)
}

// collectUsers counts the distinct users with a login session in
// run/utmp, or in the output of who when utmp cannot be read.
func collectUsers(ctx context.Context, s *session) Value[int] {
	if data, err := fs.ReadFile(s.fsys(), "run/utmp"); err == nil {
		return present(len(utmpUsers(data)))
	}
	out, err := s.run(ctx, "who")
	if err != nil {
		return Value[int]{}
	}
	users := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			users[fields[0]] = true
		}
	}
	return present(len(users))
}

// utmpUsers returns the set of user names of the login sessions in a
// utmp file. A truncated last record is ignored.
func utmpUsers(data []byte) map[string]bool {
	users := map[string]bool{}
	for len(data) >= utmpSize {
		record := data[:utmpSize]
		data = data[utmpSize:]
		if binary.LittleEndian.Uint16(record) != utmpUserProcess {
			continue
		}
		name := record[utmpUserOffset : utmpUserOffset+utmpUserSize]
		if end := bytes.IndexByte(name, 0); end >= 0 {
			name = name[:end]
		}
		if len(name) > 0 {
			users[string(name)] = true
		}
	}
	return users
}

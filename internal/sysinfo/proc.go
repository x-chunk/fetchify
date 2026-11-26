package sysinfo

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

// parentPID reads the parent process ID of pid from proc/<pid>/stat.
func parentPID(fsys fs.FS, pid int) (int, error) {
	data, err := fs.ReadFile(fsys, fmt.Sprintf("proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}
	return parseStatPPID(string(data))
}

// parseStatPPID extracts the parent process ID from the content of
// proc/<pid>/stat: "pid (comm) state ppid ...". The command name may
// contain spaces and parentheses, so the fields after it are found from
// the last closing parenthesis.
func parseStatPPID(stat string) (int, error) {
	end := strings.LastIndexByte(stat, ')')
	if end < 0 {
		return 0, fmt.Errorf("sysinfo: malformed stat %q", stat)
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 2 {
		return 0, fmt.Errorf("sysinfo: malformed stat %q", stat)
	}
	return strconv.Atoi(fields[1])
}

// processName reads the command name of pid from proc/<pid>/comm. The
// kernel truncates it to 15 bytes.
func processName(fsys fs.FS, pid int) (string, error) {
	return readTrimmed(fsys, fmt.Sprintf("proc/%d/comm", pid))
}

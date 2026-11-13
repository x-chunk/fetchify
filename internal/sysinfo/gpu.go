package sysinfo

import "strings"

// gpus returns the names of the display adapters reported by lspci, or
// nil when lspci is missing or reports none.
func gpus(run runner) []string {
	out, err := run("lspci")
	if err != nil {
		return nil
	}
	return parseLspci(string(out))
}

// parseLspci extracts the device names of VGA controllers from lspci
// output. A line looks like
// "00:02.0 VGA compatible controller: Intel Corporation UHD Graphics":
// the name is everything after the second colon.
func parseLspci(out string) []string {
	var names []string
	for line := range strings.SplitSeq(out, "\n") {
		if !strings.Contains(line, "VGA") {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		if name := strings.TrimSpace(parts[2]); name != "" {
			names = append(names, name)
		}
	}
	return names
}

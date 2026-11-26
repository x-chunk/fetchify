package ui

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// formatUptime shows days when there are any and then the two largest
// units below a day: "2d 3h 25m", "3h 25m", "5m 12s", "42s".
func formatUptime(d time.Duration) string {
	total := int64(d / time.Second)
	days, rest := total/86400, total%86400
	h, m, s := rest/3600, rest%3600/60, rest%60
	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	switch {
	case h > 0:
		parts = append(parts, fmt.Sprintf("%dh", h))
		if m > 0 {
			parts = append(parts, fmt.Sprintf("%dm", m))
		}
	case m > 0:
		parts = append(parts, fmt.Sprintf("%dm", m))
		if s > 0 && days == 0 {
			parts = append(parts, fmt.Sprintf("%ds", s))
		}
	case days == 0:
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return strings.Join(parts, " ")
}

// percentOf returns part of whole in percent, rounded and clamped to
// 0-100. A whole of 0 gives 0.
func percentOf(part, whole uint64) int {
	if whole == 0 {
		return 0
	}
	return clampPercent(float64(part) / float64(whole) * 100)
}

// clampPercent rounds p and clamps it to 0-100.
func clampPercent(p float64) int {
	if math.IsNaN(p) {
		return 0
	}
	return int(math.Round(min(max(p, 0), 100)))
}

var (
	// cpuNoise are trademark marks and words that add nothing to a CPU
	// model.
	cpuNoise = regexp.MustCompile(`\((R|TM|tm)\)|\bCPU\b|\b\d+-Core Processor\b|\bProcessor\b|@\s*[\d.]+\s*GHz\s*$`)
	// gpuNoise are company suffixes and revisions.
	gpuNoise = regexp.MustCompile(`\bCorporation\b|\bInc\.|\bCo\., Ltd\.|\(rev [0-9a-fA-F]+\)\s*$`)
	// lastBracket matches the last bracketed part of a name.
	lastBracket = regexp.MustCompile(`\[([^\]]+)\][^\[]*$`)
	// spaces matches runs of white space.
	spaces = regexp.MustCompile(`\s+`)
)

// gpuVendors maps the start of PCI vendor names to short names.
var gpuVendors = []struct{ prefix, name string }{
	{"Advanced Micro Devices", "AMD"},
	{"NVIDIA", "NVIDIA"},
	{"Intel", "Intel"},
}

// cleanCPU shortens a CPU model for display, for example
// "Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz" to "Intel Core i7-8700".
func cleanCPU(model string) string {
	return collapse(cpuNoise.ReplaceAllString(model, " "))
}

// cleanGPU shortens a GPU name of the form "<vendor> <device>". When the
// name has a bracketed part, which holds the marketing name in the PCI
// ID database, the vendor and that part are kept:
// "NVIDIA Corporation GA104 [GeForce RTX 3070]" becomes
// "NVIDIA GeForce RTX 3070".
func cleanGPU(name string) string {
	if m := lastBracket.FindStringSubmatch(name); m != nil {
		vendor := ""
		for _, v := range gpuVendors {
			if strings.HasPrefix(name, v.prefix) {
				vendor = v.name
				break
			}
		}
		if vendor == "" {
			vendor, _, _ = strings.Cut(collapse(gpuNoise.ReplaceAllString(name, " ")), " ")
		}
		if vendor != "" && !strings.HasPrefix(m[1], vendor) {
			return collapse(vendor + " " + m[1])
		}
		return collapse(m[1])
	}
	return collapse(gpuNoise.ReplaceAllString(name, " "))
}

// collapse trims s and reduces runs of white space to one space.
func collapse(s string) string {
	return strings.TrimSpace(spaces.ReplaceAllString(s, " "))
}

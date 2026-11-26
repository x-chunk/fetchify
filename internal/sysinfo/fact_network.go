package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strings"
)

// FactNetwork is the interface of the default route.
const FactNetwork Fact = "network"

// Network describes the interface that carries the default route.
type Network struct {
	// Interface is the name of the interface, for example "wlan0".
	Interface string
	// Type is "wifi", "ethernet" or "other".
	Type string
	// SSID is the name of the wireless network, or "" when the interface
	// is not wireless or the name cannot be read.
	SSID string
}

func init() {
	collectors[FactNetwork] = fact(collectNetwork, func(i *Info) *Value[Network] { return &i.Network })
}

// collectNetwork finds the default route in proc/net/route. The SSID
// comes from iw, then nmcli. The fact is Absent without a default route.
func collectNetwork(ctx context.Context, s *session) Value[Network] {
	data, err := fs.ReadFile(s.fsys(), "proc/net/route")
	if err != nil {
		return Value[Network]{}
	}
	iface := parseDefaultRoute(string(data))
	if iface == "" {
		return absent[Network]()
	}
	n := Network{Interface: iface, Type: "other"}
	dir := path.Join("sys/class/net", iface)
	switch kind, _ := readTrimmed(s.fsys(), path.Join(dir, "type")); {
	case exists(s.fsys(), path.Join(dir, "wireless")):
		n.Type = "wifi"
	case kind == "1": // ARPHRD_ETHER
		n.Type = "ethernet"
	}
	if n.Type == "wifi" {
		n.SSID = ssid(ctx, s, iface)
	}
	return present(n)
}

// parseDefaultRoute returns the interface of the first IPv4 route with
// the destination 0.0.0.0 in the format of proc/net/route, or "".
func parseDefaultRoute(data string) string {
	for line := range strings.SplitSeq(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "00000000" {
			return fields[0]
		}
	}
	return ""
}

// ssid asks iw and then nmcli for the name of the wireless network iface
// is connected to.
func ssid(ctx context.Context, s *session, iface string) string {
	if out, err := s.run(ctx, "iw", "dev", iface, "link"); err == nil {
		for line := range strings.SplitSeq(string(out), "\n") {
			if name, found := strings.CutPrefix(strings.TrimSpace(line), "SSID:"); found {
				return strings.TrimSpace(name)
			}
		}
	}
	if out, err := s.run(ctx, "nmcli", "-t", "-f", "active,ssid", "dev", "wifi"); err == nil {
		for line := range strings.SplitSeq(string(out), "\n") {
			if name, found := strings.CutPrefix(line, "yes:"); found {
				// Terse output escapes colons in values.
				return strings.ReplaceAll(name, `\:`, ":")
			}
		}
	}
	return ""
}

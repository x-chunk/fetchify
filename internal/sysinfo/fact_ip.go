package sysinfo

import (
	"context"
	"net/netip"
	"strings"
	"time"
)

// Facts about the addresses of the machine.
const (
	// FactLocalIP is the first global unicast address of the machine.
	FactLocalIP Fact = "local_ip"
	// FactPublicIP is the address the machine has on the internet.
	FactPublicIP Fact = "public_ip"
)

// DefaultPublicIPURL is the default of Options.PublicIPURL.
const DefaultPublicIPURL = "https://api.ipify.org"

// publicIPLimit bounds the response of the public IP endpoint: an IPv6
// address in text form is at most 45 bytes.
const publicIPLimit = 64

// LocalIP is an address of a local interface.
type LocalIP struct {
	// Addr is the address.
	Addr netip.Addr
	// Interface is the name of the interface that has it.
	Interface string
}

func init() {
	collectors[FactLocalIP] = fact(collectLocalIP, func(i *Info) *Value[LocalIP] { return &i.LocalIP })
	collectors[FactPublicIP] = fact(collectPublicIP, func(i *Info) *Value[netip.Addr] { return &i.PublicIP })
	factTimeouts[FactPublicIP] = time.Second
}

// collectLocalIP returns the first global unicast address of an
// interface that is up and is not a loopback. Private ranges count:
// they are what the local network knows the machine by. The fact is
// Absent when no interface has such an address.
func collectLocalIP(_ context.Context, s *session) Value[LocalIP] {
	if s.src.Interfaces == nil {
		return Value[LocalIP]{}
	}
	list, err := s.src.Interfaces()
	if err != nil {
		return Value[LocalIP]{}
	}
	for _, iface := range list {
		if !iface.Up || iface.Loopback {
			continue
		}
		for _, addr := range iface.Addrs {
			if addr.IsGlobalUnicast() {
				return present(LocalIP{Addr: addr.Unmap(), Interface: iface.Name})
			}
		}
	}
	return absent[LocalIP]()
}

// collectPublicIP asks an HTTPS endpoint for the address the machine
// appears from. Nothing is sent besides the request itself, and only a
// response that parses as an address is accepted.
func collectPublicIP(ctx context.Context, s *session) Value[netip.Addr] {
	if s.src.Fetch == nil {
		return Value[netip.Addr]{}
	}
	url := s.opts.PublicIPURL
	if url == "" {
		url = DefaultPublicIPURL
	}
	body, err := s.src.Fetch(ctx, url, publicIPLimit)
	if err != nil {
		return Value[netip.Addr]{}
	}
	addr, err := netip.ParseAddr(strings.TrimSpace(string(body)))
	if err != nil {
		return Value[netip.Addr]{}
	}
	return present(addr.Unmap())
}

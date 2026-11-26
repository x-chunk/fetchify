package sysinfo

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"
)

func TestLocalIP(t *testing.T) {
	list := []NetInterface{
		{Name: "lo", Up: true, Loopback: true, Addrs: []netip.Addr{netip.MustParseAddr("127.0.0.1")}},
		{Name: "eth0", Up: false, Addrs: []netip.Addr{netip.MustParseAddr("10.0.0.5")}},
		{Name: "wlan0", Up: true, Addrs: []netip.Addr{
			netip.MustParseAddr("fe80::1"),
			netip.MustParseAddr("192.168.1.23"),
		}},
	}
	tests := []struct {
		name string
		src  Source
		want Value[LocalIP]
	}{
		{"first usable", Source{Interfaces: func() ([]NetInterface, error) { return list, nil }},
			present(LocalIP{Addr: netip.MustParseAddr("192.168.1.23"), Interface: "wlan0"})},
		{"offline", Source{Interfaces: func() ([]NetInterface, error) { return list[:2], nil }}, absent[LocalIP]()},
		{"error", Source{Interfaces: func() ([]NetInterface, error) { return nil, errors.New("x") }}, Value[LocalIP]{}},
		{"no source", Source{}, Value[LocalIP]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactLocalIP}, Options{}).LocalIP
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPublicIP(t *testing.T) {
	answer := func(body string, err error) func(context.Context, string, int64) ([]byte, error) {
		return func(context.Context, string, int64) ([]byte, error) { return []byte(body), err }
	}
	tests := []struct {
		name  string
		fetch func(context.Context, string, int64) ([]byte, error)
		want  Value[netip.Addr]
	}{
		{"ipv4", answer("203.0.113.7\n", nil), present(netip.MustParseAddr("203.0.113.7"))},
		{"ipv6", answer("2001:db8::1", nil), present(netip.MustParseAddr("2001:db8::1"))},
		{"html", answer("<html>blocked</html>", nil), Value[netip.Addr]{}},
		{"offline", answer("", errors.New("no route")), Value[netip.Addr]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), Source{Fetch: tt.fetch}, []Fact{FactPublicIP}, Options{}).PublicIP
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPublicIPRequest(t *testing.T) {
	var gotURL string
	var gotLimit int64
	src := Source{Fetch: func(_ context.Context, url string, limit int64) ([]byte, error) {
		gotURL, gotLimit = url, limit
		return []byte("1.2.3.4"), nil
	}}
	Collect(context.Background(), src, []Fact{FactPublicIP}, Options{})
	if gotURL != DefaultPublicIPURL || gotLimit != publicIPLimit {
		t.Errorf("default request = %q, %d", gotURL, gotLimit)
	}
	Collect(context.Background(), src, []Fact{FactPublicIP}, Options{PublicIPURL: "https://example.test/ip"})
	if gotURL != "https://example.test/ip" {
		t.Errorf("configured request = %q", gotURL)
	}
	if got := (Options{Timeout: time.Millisecond}).timeout(FactPublicIP); got != time.Second {
		t.Errorf("public_ip timeout = %v, want its own default of 1s", got)
	}
}

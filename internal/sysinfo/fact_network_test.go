package sysinfo

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestNetwork(t *testing.T) {
	local := "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\n" +
		"wlan0\t0001A8C0\t00000000\t0001\t0\t0\t600\t00FFFFFF\n"
	route := local + "wlan0\t00000000\t0101A8C0\t0003\t0\t0\t600\t00000000\n"
	fsys := func(extra fstest.MapFS) fstest.MapFS {
		m := fstest.MapFS{"proc/net/route": {Data: []byte(route)}}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	wifi := fstest.MapFS{"sys/class/net/wlan0/wireless/x": {}, "sys/class/net/wlan0/type": {Data: []byte("1\n")}}
	run := func(outputs map[string]string) func(context.Context, string, ...string) ([]byte, error) {
		return machine{commands: outputs}.source().Run
	}
	tests := []struct {
		name string
		src  Source
		want Value[Network]
	}{
		{"wifi through iw", Source{FS: fsys(wifi), Run: run(map[string]string{"iw dev wlan0 link": "Connected to aa:bb\n\tSSID: Home Net\n"})},
			present(Network{Interface: "wlan0", Type: "wifi", SSID: "Home Net"})},
		{"wifi through nmcli", Source{FS: fsys(wifi), Run: run(map[string]string{"nmcli -t -f active,ssid dev wifi": "no:Other\nyes:Cafe\\: Free\n"})},
			present(Network{Interface: "wlan0", Type: "wifi", SSID: "Cafe: Free"})},
		{"wifi without tools", Source{FS: fsys(wifi)}, present(Network{Interface: "wlan0", Type: "wifi"})},
		{"ethernet", Source{FS: fsys(fstest.MapFS{"sys/class/net/wlan0/type": {Data: []byte("1\n")}})},
			present(Network{Interface: "wlan0", Type: "ethernet"})},
		{"other", Source{FS: fsys(nil)}, present(Network{Interface: "wlan0", Type: "other"})},
		{"no default route", Source{FS: fstest.MapFS{"proc/net/route": {Data: []byte(local)}}}, absent[Network]()},
		{"no proc", Source{}, Value[Network]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactNetwork}, Options{}).Network
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBluetooth(t *testing.T) {
	file := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s + "\n")} }
	rfkill := func(soft string) fstest.MapFS {
		return fstest.MapFS{
			"sys/class/rfkill/rfkill0/type": file("bluetooth"),
			"sys/class/rfkill/rfkill0/soft": file(soft),
			"sys/class/rfkill/rfkill0/hard": file("0"),
			"sys/class/rfkill/rfkill1/type": file("wlan"),
			"sys/class/rfkill/rfkill1/soft": file("1"),
		}
	}
	on := rfkill("0")
	on["sys/class/bluetooth/hci0/x"] = &fstest.MapFile{}
	tests := []struct {
		name string
		fs   fstest.MapFS
		want Value[bool]
	}{
		{"on", on, present(true)},
		{"blocked", rfkill("1"), present(false)},
		{"no adapter", fstest.MapFS{}, absent[bool]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), Source{FS: tt.fs}, []Fact{FactBluetooth}, Options{}).Bluetooth
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

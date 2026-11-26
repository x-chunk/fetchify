package sysinfo

import (
	"context"
	"reflect"
	"testing"
)

func TestUnescapeMount(t *testing.T) {
	tests := map[string]string{
		`/mnt/My\040Drive`: "/mnt/My Drive",
		`/a\134b`:          `/a\b`,
		`/bad\04`:          `/bad\04`,
		`/bad\999x`:        `/bad\999x`,
		`/plain`:           "/plain",
	}
	for in, want := range tests {
		if got := unescapeMount(in); got != want {
			t.Errorf("unescapeMount(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsIgnoredMount(t *testing.T) {
	tests := []struct {
		mount string
		want  bool
	}{
		{"/", false},
		{"/home", false},
		{"/boot", true},
		{"/boot/efi", true},
		{"/var/lib/docker", true},
		{"/variable", false},
		{"/opt/run", false},
		{"/mnt/usb", false},
		{"/nix/store", true},
		{"/nix", false},
	}
	for _, tt := range tests {
		if got := isIgnoredMount(tt.mount, DefaultDiskIgnore); got != tt.want {
			t.Errorf("isIgnoredMount(%q) = %v, want %v", tt.mount, got, tt.want)
		}
	}
	if !isIgnoredMount("/x/y", []string{"/x/"}) || isIgnoredMount("/x", []string{""}) {
		t.Error("trailing slash or empty entry handled wrongly")
	}
}

func TestDisks(t *testing.T) {
	disk := func(point, fstype string, used, total, avail uint64) Disk {
		return Disk{Mountpoint: point, Filesystem: fstype, Usage: Usage{Used: used * gib, Total: total * gib, Available: avail * gib}}
	}
	tests := []struct {
		name    string
		machine machine
		opts    Options
		want    Value[[]Disk]
	}{
		{"desktop", ubuntuDesktop, Options{}, present([]Disk{
			disk("/", "ext4", 100, 500, 376),
			disk("/home", "ext4", 250, 1000, 700),
			disk("/mnt/My Drive", "exfat", 1, 64, 63),
		})},
		{"btrfs subvolumes and bind mounts", btrfsBox, Options{}, present([]Disk{
			disk("/", "btrfs", 300, 1000, 700),
			disk("/data", "xfs", 10, 100, 90),
		})},
		{"own ignore list", ubuntuDesktop, Options{DiskIgnore: []string{"/home", "/mnt"}}, present([]Disk{
			disk("/", "ext4", 100, 500, 376),
			disk("/boot/efi", "vfat", 0, 1, 1),
		})},
		{"container", container, Options{}, absent[[]Disk]()},
		{"no mount table", machine{}, Options{}, Value[[]Disk]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.machine.source(), []Fact{FactDisks}, tt.opts).Disks
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

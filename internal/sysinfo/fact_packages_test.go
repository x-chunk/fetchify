package sysinfo

import (
	"context"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestCountDpkg(t *testing.T) {
	status := "Package: a\nStatus: install ok installed\n\n" +
		"Package: b\nStatus: deinstall ok config-files\n\n" +
		"Package: c\nStatus: purge ok not-installed\n\n" +
		"Package: d\nStatus: install ok installed\nDescription: Status: installed\n"
	if got := countDpkg([]byte(status)); got != 2 {
		t.Errorf("countDpkg = %d, want 2", got)
	}
}

func TestPackages(t *testing.T) {
	dir := &fstest.MapFile{Mode: fs.ModeDir | 0o755}
	tests := []struct {
		name string
		src  Source
		want Value[[]PackageCount]
	}{
		{"ubuntu", ubuntuDesktop.source(), present([]PackageCount{{"dpkg", 3}, {"flatpak", 2}, {"snap", 2}})},
		{"arch", archLaptop.source(), present([]PackageCount{{"pacman", 4}, {"flatpak", 1}})},
		{"alpine", container.source(), present([]PackageCount{{"apk", 3}})},
		{"fedora with rpm", btrfsBox.source(), present([]PackageCount{{"rpm", 3}})},
		{"nothing", machine{}.source(), absent[[]PackageCount]()},
		{"portage", Source{FS: fstest.MapFS{
			"var/db/pkg/sys-apps/portage-3.0/CONTENTS":   {},
			"var/db/pkg/sys-apps/coreutils-9.5/CONTENTS": {},
			"var/db/pkg/dev-lang/go-1.23/CONTENTS":       {},
			"var/db/pkg/.keep":                           {},
		}}, present([]PackageCount{{"portage", 3}})},
		{"xbps and nix", Source{
			FS: fstest.MapFS{"var/db/xbps": dir, "nix/store": dir},
			Run: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				return map[string][]byte{"xbps-query": []byte("ii a\nii b\n"), "nix-store": []byte("/nix/store/x\n")}[name], nil
			},
		}, present([]PackageCount{{"xbps", 2}, {"nix", 1}})},
		{"failing tool is dropped", Source{FS: fstest.MapFS{"var/lib/rpm": dir}}, absent[[]PackageCount]()},
		{"empty database is dropped", Source{FS: fstest.MapFS{"var/lib/pacman/local": dir}}, absent[[]PackageCount]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collect(context.Background(), tt.src, []Fact{FactPackages}, Options{}).Packages
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

package distro

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseReleaseFiles(t *testing.T) {
	tests := []struct {
		file string
		want Release
	}{
		{"ubuntu", Release{ID: "ubuntu", IDLike: []string{"debian"}, Name: "Ubuntu", PrettyName: "Ubuntu 24.04.1 LTS", VersionID: "24.04"}},
		{"arch", Release{ID: "arch", Name: "Arch Linux", PrettyName: "Arch Linux", BuildID: "rolling"}},
		{"fedora", Release{ID: "fedora", Name: "Fedora Linux", PrettyName: "Fedora Linux 41 (Workstation Edition)", VersionID: "41", VariantID: "workstation"}},
		{"linuxmint", Release{ID: "linuxmint", IDLike: []string{"ubuntu", "debian"}, Name: "Linux Mint", PrettyName: "Linux Mint 22", VersionID: "22"}},
		{"opensuse-tumbleweed", Release{ID: "opensuse-tumbleweed", IDLike: []string{"opensuse", "suse"}, Name: "openSUSE Tumbleweed", PrettyName: "openSUSE Tumbleweed", VersionID: "20241120"}},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "os-release", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			if got := ParseRelease(data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestParseReleaseSyntax(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Release
	}{
		{"empty", "", Release{}},
		{"double quotes", `NAME="Foo Linux"`, Release{Name: "Foo Linux"}},
		{"single quotes", `NAME='Foo "Linux"'`, Release{Name: `Foo "Linux"`}},
		{"escapes", `PRETTY_NAME="Foo \"Bar\" \$5 \\ x"`, Release{PrettyName: `Foo "Bar" $5 \ x`}},
		{"unquoted escape", `NAME=Foo\ Linux`, Release{Name: "Foo Linux"}},
		{"comment", "# ID=nope\nID=yes", Release{ID: "yes"}},
		{"lower case", "ID=Foo\nID_LIKE=\"Bar  BAZ\"", Release{ID: "foo", IDLike: []string{"bar", "baz"}}},
		{"last wins", "ID=a\nID=b", Release{ID: "b"}},
		{"spaces and CRLF", "  ID = x \r\n", Release{ID: "x"}},
		{"unterminated", "NAME=\"Foo\nID=x", Release{ID: "x"}},
		{"no equals", "garbage\nID=x", Release{ID: "x"}},
		{"empty value", `NAME=""`, Release{}},
		{"binary junk", "\x00\xff\xfe=\x01\nID=x\x00", Release{ID: "x\x00"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRelease([]byte(tt.in)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func FuzzParseRelease(f *testing.F) {
	f.Add([]byte("ID=\"x\\\"\"\nNAME='y'"))
	f.Fuzz(func(t *testing.T, data []byte) {
		ParseRelease(data)
	})
}

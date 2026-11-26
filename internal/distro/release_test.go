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

// TestLookupReleaseFiles parses the os-release file of every distribution
// in the table and checks which entry it resolves to.
func TestLookupReleaseFiles(t *testing.T) {
	want := map[string]string{
		"almalinux": "almalinux", "alpine": "alpine", "arch": "arch", "artix": "artix",
		"cachyos": "cachyos", "centos": "centos", "debian": "debian", "elementary": "elementary",
		"endeavouros": "endeavouros", "fedora": "fedora", "gentoo": "gentoo", "kali": "kali",
		"linuxmint": "linuxmint", "manjaro": "manjaro", "nixos": "nixos", "ol": "rhel",
		"opensuse-tumbleweed": "opensuse", "pop": "pop", "raspbian": "debian", "rhel": "rhel",
		"rocky": "rocky", "sles": "opensuse", "ubuntu": "ubuntu", "void": "void", "zorin": "zorin",
	}
	files, err := os.ReadDir(filepath.Join("testdata", "os-release"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(want) {
		t.Errorf("%d fixtures, %d expectations", len(files), len(want))
	}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join("testdata", "os-release", f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if got := Lookup(ParseRelease(data)).ID; got != want[f.Name()] {
			t.Errorf("%s resolves to %q, want %q", f.Name(), got, want[f.Name()])
		}
	}
}

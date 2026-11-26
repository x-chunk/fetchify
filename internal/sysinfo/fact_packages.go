package sysinfo

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"path"
	"strings"
)

// FactPackages is the number of installed packages per package manager.
const FactPackages Fact = "packages"

// PackageCount is the number of packages one manager has installed.
type PackageCount struct {
	// Manager is the name of the package manager, for example "dpkg".
	Manager string
	// Count is the number of installed packages, always above 0.
	Count int
}

func init() {
	collectors[FactPackages] = fact(collectPackages, func(i *Info) *Value[[]PackageCount] { return &i.Packages })
}

// packageManager counts the packages of one manager. It returns false
// when the manager is not installed or its count cannot be determined.
type packageManager struct {
	name  string
	count func(ctx context.Context, s *session) (int, bool)
}

// packageManagers lists the managers in the order they are reported.
// Counting reads the database where its format is simple, which is faster
// than running the tool and needs no time budget.
var packageManagers = []packageManager{
	{"pacman", func(_ context.Context, s *session) (int, bool) {
		return countDirs(s.fsys(), "var/lib/pacman/local")
	}},
	{"dpkg", func(_ context.Context, s *session) (int, bool) {
		data, err := fs.ReadFile(s.fsys(), "var/lib/dpkg/status")
		if err != nil {
			return 0, false
		}
		return countDpkg(data), true
	}},
	{"rpm", func(ctx context.Context, s *session) (int, bool) {
		if !exists(s.fsys(), "var/lib/rpm") && !exists(s.fsys(), "usr/lib/sysimage/rpm") {
			return 0, false
		}
		return s.countLines(ctx, "rpm", "-qa")
	}},
	{"xbps", func(ctx context.Context, s *session) (int, bool) {
		if !exists(s.fsys(), "var/db/xbps") {
			return 0, false
		}
		return s.countLines(ctx, "xbps-query", "-l")
	}},
	{"apk", func(_ context.Context, s *session) (int, bool) {
		data, err := fs.ReadFile(s.fsys(), "lib/apk/db/installed")
		if err != nil {
			return 0, false
		}
		return countPrefixed(data, "P:"), true
	}},
	{"portage", func(_ context.Context, s *session) (int, bool) {
		return countPortage(s.fsys())
	}},
	{"nix", func(ctx context.Context, s *session) (int, bool) {
		if !exists(s.fsys(), "nix/store") {
			return 0, false
		}
		return s.countLines(ctx, "nix-store", "-q", "--requisites", "/run/current-system/sw")
	}},
	{"flatpak", func(_ context.Context, s *session) (int, bool) {
		system, okSystem := countDirs(s.fsys(), "var/lib/flatpak/app")
		var user int
		var okUser bool
		if home := s.home(); home != "" {
			user, okUser = countDirs(s.fsys(), path.Join(home, ".local/share/flatpak/app"))
		}
		return system + user, okSystem || okUser
	}},
	{"snap", func(_ context.Context, s *session) (int, bool) {
		entries, err := fs.ReadDir(s.fsys(), "snap")
		if err != nil {
			return 0, false
		}
		n := 0
		for _, e := range entries {
			if e.IsDir() && e.Name() != "bin" && e.Name() != "README" {
				n++
			}
		}
		return n, true
	}},
}

// collectPackages counts the packages of every manager whose database
// exists. Managers with no packages, and managers whose tool fails or
// times out, are left out rather than reported as 0. The fact is Absent
// when no manager is left.
func collectPackages(ctx context.Context, s *session) Value[[]PackageCount] {
	var counts []PackageCount
	for _, m := range packageManagers {
		if n, ok := m.count(ctx, s); ok && n > 0 {
			counts = append(counts, PackageCount{Manager: m.name, Count: n})
		}
	}
	if len(counts) == 0 {
		return absent[[]PackageCount]()
	}
	return present(counts)
}

// countDpkg counts the paragraphs of a dpkg status file whose state, the
// last word of the Status line, is "installed". Packages that were
// removed but kept their configuration are "config-files".
func countDpkg(data []byte) int {
	n := 0
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		status, found := bytes.CutPrefix(line, []byte("Status:"))
		if !found {
			continue
		}
		fields := bytes.Fields(status)
		if len(fields) > 0 && string(fields[len(fields)-1]) == "installed" {
			n++
		}
	}
	return n
}

// countPrefixed counts the lines of data that start with prefix.
func countPrefixed(data []byte, prefix string) int {
	n := 0
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if bytes.HasPrefix(line, []byte(prefix)) {
			n++
		}
	}
	return n
}

// countNonEmptyLines counts the lines of data that are not blank.
func countNonEmptyLines(data []byte) int {
	n := 0
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			n++
		}
	}
	return n
}

// countPortage counts the package directories two levels under
// var/db/pkg, which holds one directory per category.
func countPortage(fsys fs.FS) (int, bool) {
	categories, err := fs.ReadDir(fsys, "var/db/pkg")
	if err != nil {
		return 0, false
	}
	total := 0
	for _, c := range categories {
		if !c.IsDir() {
			continue
		}
		if n, ok := countDirs(fsys, path.Join("var/db/pkg", c.Name())); ok {
			total += n
		}
	}
	return total, true
}

// countDirs counts the directories directly inside dir. It reports false
// when dir cannot be read.
func countDirs(fsys fs.FS, dir string) (int, bool) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return 0, false
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			n++
		}
	}
	return n, true
}

// exists reports whether name exists in fsys.
func exists(fsys fs.FS, name string) bool {
	_, err := fs.Stat(fsys, name)
	return err == nil
}

// errNoRunner is returned when a Source cannot run commands.
var errNoRunner = errors.New("sysinfo: source cannot run commands")

// run runs a command through the Source.
func (s *session) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if s.src.Run == nil {
		return nil, errNoRunner
	}
	return s.src.Run(ctx, name, args...)
}

// countLines runs a command and counts the non-blank lines of its output.
func (s *session) countLines(ctx context.Context, name string, args ...string) (int, bool) {
	out, err := s.run(ctx, name, args...)
	if err != nil {
		return 0, false
	}
	return countNonEmptyLines(out), true
}

// home returns the home directory from $HOME as a path for the Source's
// file system, without the leading slash, or "" when it is unset or not
// absolute.
func (s *session) home() string {
	home := s.getenv("HOME")
	if !strings.HasPrefix(home, "/") {
		return ""
	}
	return strings.TrimPrefix(path.Clean(home), "/")
}

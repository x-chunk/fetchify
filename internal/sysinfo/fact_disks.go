package sysinfo

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
)

// FactDisks is the use of every mounted file system worth showing.
const FactDisks Fact = "disks"

// Disk is a mounted file system.
type Disk struct {
	// Mountpoint is the path the file system is mounted at.
	Mountpoint string
	// Filesystem is the file system type, for example "ext4".
	Filesystem string
	// Usage is the space in use. Used counts the blocks reserved for
	// root, Available does not.
	Usage Usage
}

// DefaultDiskIgnore is the default of Options.DiskIgnore: system
// plumbing, boot partitions, caches and package images. /etc holds the
// files a container runtime bind-mounts from the host.
var DefaultDiskIgnore = []string{
	"/boot", "/efi", "/etc", "/run", "/var", "/tmp", "/dev", "/sys", "/proc", "/snap", "/nix/store",
}

// pseudoFilesystems are file system types that hold no user data: kernel
// interfaces, memory-backed file systems and read-only images.
var pseudoFilesystems = map[string]bool{
	"proc": true, "sysfs": true, "tmpfs": true, "devtmpfs": true, "devpts": true,
	"cgroup": true, "cgroup2": true, "overlay": true, "squashfs": true, "efivarfs": true,
	"fuse.portal": true, "autofs": true, "ramfs": true, "debugfs": true, "tracefs": true,
	"securityfs": true, "pstore": true, "bpf": true, "configfs": true, "mqueue": true,
	"hugetlbfs": true, "binfmt_misc": true, "nsfs": true, "fusectl": true, "rpc_pipefs": true,
}

func init() {
	collectors[FactDisks] = fact(collectDisks, func(i *Info) *Value[[]Disk] { return &i.Disks })
}

// mount is an entry of the mount table.
type mount struct {
	device, point, fstype string
}

// collectDisks lists the file systems that hold user data, in mount
// order. Mount points inside the ignore list are left out, as are file
// systems that cannot be measured or have a size of 0. Several mounts of
// one device, such as btrfs subvolumes and bind mounts, are reported once
// at the shortest mount point. The fact is Absent when nothing is left.
func collectDisks(_ context.Context, s *session) Value[[]Disk] {
	data, err := fs.ReadFile(s.fsys(), "proc/self/mounts")
	if err != nil || s.src.Statfs == nil {
		return Value[[]Disk]{}
	}
	ignore := s.opts.DiskIgnore
	if ignore == nil {
		ignore = DefaultDiskIgnore
	}
	var mounts []mount
	for _, m := range parseMounts(string(data)) {
		if !pseudoFilesystems[m.fstype] && !isIgnoredMount(m.point, ignore) {
			mounts = append(mounts, m)
		}
	}
	var disks []Disk
	for _, m := range shortestPerDevice(mounts) {
		st, err := s.src.Statfs(m.point)
		if err != nil || st.Total == 0 {
			continue
		}
		free := min(st.Free, st.Total)
		disks = append(disks, Disk{
			Mountpoint: m.point,
			Filesystem: m.fstype,
			Usage:      Usage{Used: st.Total - free, Total: st.Total, Available: min(st.Available, free)},
		})
	}
	if len(disks) == 0 {
		return absent[[]Disk]()
	}
	return present(disks)
}

// parseMounts parses the mount table in the format of proc/self/mounts:
// "device mountpoint type options dump pass", with spaces and other
// special characters in the first two fields escaped as octal.
func parseMounts(data string) []mount {
	var mounts []mount
	for line := range strings.SplitSeq(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		mounts = append(mounts, mount{
			device: unescapeMount(fields[0]),
			point:  unescapeMount(fields[1]),
			fstype: fields[2],
		})
	}
	return mounts
}

// unescapeMount decodes the octal escapes of the mount table, such as
// "\040" for a space. Malformed escapes are kept as they are.
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+4 <= len(s) {
			if n, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(n))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// isIgnoredMount reports whether point is one of ignore or lies inside
// one of them. Matching is by path component, so "/variable" is not
// ignored because of "/var".
func isIgnoredMount(point string, ignore []string) bool {
	for _, dir := range ignore {
		dir = strings.TrimSuffix(dir, "/")
		if dir == "" {
			continue
		}
		if point == dir || strings.HasPrefix(point, dir+"/") {
			return true
		}
	}
	return false
}

// shortestPerDevice keeps, for every device, the mount with the shortest
// mount point, and returns the kept mounts in their original order.
func shortestPerDevice(mounts []mount) []mount {
	best := map[string]int{} // device -> index into mounts
	for i, m := range mounts {
		if j, ok := best[m.device]; !ok || len(m.point) < len(mounts[j].point) {
			best[m.device] = i
		}
	}
	var kept []mount
	for i, m := range mounts {
		if best[m.device] == i {
			kept = append(kept, m)
		}
	}
	return kept
}

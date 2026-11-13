package sysinfo

import (
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

// ignoredMounts lists the directory trees whose file systems are system
// plumbing or temporary media rather than storage the user cares about.
var ignoredMounts = []string{"/boot", "/efi", "/mnt", "/run", "/var", "/tmp", "/dev", "/sys", "/proc"}

// disks returns the mounted physical file systems outside ignoredMounts.
// File systems whose usage cannot be read are skipped.
func disks() []Disk {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	var result []Disk
	for _, partition := range partitions {
		if isIgnoredMount(partition.Mountpoint) {
			continue
		}
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}
		result = append(result, Disk{
			Usage:      Usage{Used: usage.Used, Total: usage.Total, Percent: usage.UsedPercent},
			Mountpoint: partition.Mountpoint,
			Filesystem: partition.Fstype,
		})
	}
	return result
}

// isIgnoredMount reports whether mount is one of ignoredMounts or lies
// inside one of them. Matching is by path component, so "/variable" is
// not ignored because of "/var".
func isIgnoredMount(mount string) bool {
	for _, ignored := range ignoredMounts {
		if mount == ignored || strings.HasPrefix(mount, ignored+"/") {
			return true
		}
	}
	return false
}

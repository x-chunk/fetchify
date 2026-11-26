package sysinfo

import "syscall"

// statfs reads the sizes of the file system mounted at path.
func statfs(path string) (FSStat, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return FSStat{}, err
	}
	size := uint64(st.Bsize)
	return FSStat{
		Total:     st.Blocks * size,
		Free:      st.Bfree * size,
		Available: st.Bavail * size,
	}, nil
}

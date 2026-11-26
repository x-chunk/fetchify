package sysinfo

import (
	"io/fs"
	"strings"
)

// readTrimmed returns the content of the file at name in fsys without
// surrounding white space. sysfs and procfs files end with a newline.
func readTrimmed(fsys fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// getenv calls s.Getenv, or returns "" when the Source has none.
func (s *session) getenv(key string) string {
	if s.src.Getenv == nil {
		return ""
	}
	return s.src.Getenv(key)
}

// fsys returns the file system of the Source, or an empty one when the
// Source has none, so collectors can read without a nil check.
func (s *session) fsys() fs.FS {
	if s.src.FS == nil {
		return emptyFS{}
	}
	return s.src.FS
}

// emptyFS is a file system without files.
type emptyFS struct{}

// Open implements fs.FS.
func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

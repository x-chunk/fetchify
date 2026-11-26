// Package golden compares test output with files stored in testdata.
//
// Golden files hold raw bytes, including escape sequences. Importing the
// package registers the -update flag, which rewrites the files instead of
// comparing them. Review the resulting diff by hand before committing it.
package golden

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files in testdata")

// Assert compares got with the content of testdata/<name>.golden relative
// to the test's package directory, which is the working directory of
// go test. A mismatch or a missing file fails t without stopping it.
// With the -update flag Assert writes got to the file instead.
func Assert(t testing.TB, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if err := check(path, got, *update); err != nil {
		t.Error(err)
	}
}

// check holds the logic of Assert without the dependency on testing.TB,
// so its failures can be tested.
func check(path string, got []byte, update bool) error {
	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, got, 0o644)
	}
	want, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("golden file %s is missing; run the test with -update to create it", path)
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("output differs from %s\n got: %q\nwant: %q", path, got, want)
	}
	return nil
}

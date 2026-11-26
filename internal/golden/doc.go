// Package golden compares test output with files stored in testdata.
//
// Golden files hold raw bytes, including escape sequences. Importing the
// package registers the -update flag, which rewrites the files instead of
// comparing them. Review the resulting diff by hand before committing it.
package golden

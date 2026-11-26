package config

import (
	"reflect"
	"testing"
)

func TestDefaultsAreIndependent(t *testing.T) {
	a := Defaults()
	a.Groups[0].Items[0] = "changed"
	a.Items["disks"].Ignore[0] = "changed"
	if b := Defaults(); reflect.DeepEqual(a, b) {
		t.Error("Defaults returns shared slices")
	}
}

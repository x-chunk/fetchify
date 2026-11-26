package config

import (
	"reflect"
	"testing"
)

// TestDefaultTOML keeps the documented file and the code in sync.
func TestDefaultTOML(t *testing.T) {
	cfg, err := Parse("default.toml", DefaultTOML(), known)
	if err != nil {
		t.Fatal(err)
	}
	if want := Defaults(); !reflect.DeepEqual(cfg, want) {
		t.Errorf("default.toml decodes to\n%+v\nwant\n%+v", cfg, want)
	}
}

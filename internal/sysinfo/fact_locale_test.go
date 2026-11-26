package sysinfo

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
	"time"
)

func TestLocale(t *testing.T) {
	tests := []struct {
		env  map[string]string
		want Value[string]
	}{
		{map[string]string{"LANG": "en_US.UTF-8", "LC_MESSAGES": "de_DE.UTF-8"}, present("de_DE.UTF-8")},
		{map[string]string{"LANG": "en_US.UTF-8", "LC_ALL": "C.UTF-8"}, present("C.UTF-8")},
		{map[string]string{"LANG": "ru_RU.UTF-8"}, present("ru_RU.UTF-8")},
		{nil, absent[string]()},
	}
	for _, tt := range tests {
		if got := collectOne(t, machine{env: tt.env}, FactLocale).Locale; got != tt.want {
			t.Errorf("%v: got %+v, want %+v", tt.env, got, tt.want)
		}
	}
}

func TestTimezone(t *testing.T) {
	link := func(target string, err error) func(string) (string, error) {
		return func(string) (string, error) { return target, err }
	}
	env := func(vars map[string]string) func(string) string { return func(k string) string { return vars[k] } }
	tests := []struct {
		name string
		src  Source
		want Value[string]
	}{
		{"TZ", Source{Getenv: env(map[string]string{"TZ": ":Asia/Almaty"})}, present("Asia/Almaty")},
		{"TZ as a file", Source{Getenv: env(map[string]string{"TZ": "/etc/localtime"}), Readlink: link("../usr/share/zoneinfo/UTC", nil)}, present("UTC")},
		{"localtime link", Source{Readlink: link("/usr/share/zoneinfo/Europe/Berlin", nil)}, present("Europe/Berlin")},
		{"etc/timezone", Source{Readlink: link("", errors.New("not a link")), FS: fstest.MapFS{"etc/timezone": {Data: []byte("America/New_York\n")}}}, present("America/New_York")},
		{"nothing", Source{}, Value[string]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Collect(context.Background(), tt.src, []Fact{FactTimezone}, Options{}).Timezone; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDate(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	src := Source{Now: func() time.Time { return now }}
	if got := Collect(context.Background(), src, []Fact{FactDate}, Options{}).Date; got != present(now) {
		t.Errorf("got %+v", got)
	}
	if got := Collect(context.Background(), Source{}, []Fact{FactDate}, Options{}).Date; got.State != Unknown {
		t.Errorf("without a clock: %+v", got)
	}
}

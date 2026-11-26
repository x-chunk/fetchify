package sysinfo

import (
	"context"
	"io/fs"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

// sequenceFS serves proc/stat from a list of contents, one per open.
type sequenceFS struct {
	mu       sync.Mutex
	contents []string
}

func (s *sequenceFS) Open(name string) (fs.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if name != "proc/stat" || len(s.contents) == 0 {
		return nil, fs.ErrNotExist
	}
	data := s.contents[0]
	s.contents = s.contents[1:]
	return fstest.MapFS{"f": {Data: []byte(data)}}.Open("f")
}

func TestCPUUsage(t *testing.T) {
	defer func(d time.Duration) { cpuUsageInterval = d }(cpuUsageInterval)
	cpuUsageInterval = time.Millisecond
	tests := []struct {
		name     string
		contents []string
		want     Value[float64]
	}{
		{"quarter busy", []string{
			"cpu  100 0 100 700 100 0 0 0 50 0\ncpu0 1 2 3 4\n",
			"cpu  150 0 100 850 100 0 0 0 90 0\n",
		}, present(25.0)},
		{"no time passed", []string{"cpu 1 1 1 1 1\n", "cpu 1 1 1 1 1\n"}, Value[float64]{}},
		{"counters reset", []string{"cpu 5 5 5 5 5\n", "cpu 1 1 1 1 1\n"}, Value[float64]{}},
		{"one sample", []string{"cpu 1 1 1 1 1\n"}, Value[float64]{}},
		{"garbage", []string{"intr 1\n", "intr 2\n"}, Value[float64]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := Source{FS: &sequenceFS{contents: tt.contents}}
			got := Collect(context.Background(), src, []Fact{FactCPUUsage}, Options{}).CPUUsage
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCPUUsageHonorsDeadline(t *testing.T) {
	defer func(d time.Duration) { cpuUsageInterval = d }(cpuUsageInterval)
	cpuUsageInterval = time.Hour
	src := Source{FS: &sequenceFS{contents: []string{"cpu 1 1 1 1 1\n", "cpu 2 2 2 2 2\n"}}}
	start := time.Now()
	got := Collect(context.Background(), src, []Fact{FactCPUUsage}, Options{Timeout: 10 * time.Millisecond}).CPUUsage
	if got.State != Unknown || time.Since(start) > time.Second {
		t.Errorf("got %+v after %v", got, time.Since(start))
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		data string
		want Value[[3]float64]
	}{
		{"0.52 0.58 0.59 1/1234 5678\n", present([3]float64{0.52, 0.58, 0.59})},
		{"0.52 x 0.59", Value[[3]float64]{}},
		{"1 2", Value[[3]float64]{}},
	}
	for _, tt := range tests {
		src := Source{FS: fstest.MapFS{"proc/loadavg": {Data: []byte(tt.data)}}}
		if got := Collect(context.Background(), src, []Fact{FactLoad}, Options{}).Load; got != tt.want {
			t.Errorf("%q: got %+v, want %+v", tt.data, got, tt.want)
		}
	}
}

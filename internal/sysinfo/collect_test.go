package sysinfo

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// into returns a collector that stores the result of get in *dst instead
// of a field of Info. collect runs the store on its own goroutine, like
// for a real field.
func into(dst *Value[string], get func(context.Context, *session) Value[string]) collector {
	return func(ctx context.Context, s *session) func(*Info) {
		v := get(ctx, s)
		return func(*Info) { *dst = v }
	}
}

// constant returns a collector that stores v in *dst and counts its calls.
func constant(dst *Value[string], v string, calls *atomic.Int32) collector {
	return into(dst, func(context.Context, *session) Value[string] {
		calls.Add(1)
		return present(v)
	})
}

func TestCollectRequestedOnly(t *testing.T) {
	var a, b atomic.Int32
	var gotA, gotB Value[string]
	registry := map[Fact]collector{"a": constant(&gotA, "A", &a), "b": constant(&gotB, "B", &b)}
	collect(context.Background(), Source{}, []Fact{"a", "a", "nope"}, Options{}, registry)
	if gotA != present("A") || gotB.State != Unknown {
		t.Errorf("a = %+v, b = %+v; want A and unknown", gotA, gotB)
	}
	if a.Load() != 1 || b.Load() != 0 {
		t.Errorf("calls a=%d b=%d, want 1 and 0", a.Load(), b.Load())
	}
}

func TestCollectTimeout(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	var got Value[string]
	blocking := into(&got, func(context.Context, *session) Value[string] {
		<-release // ignores the context on purpose
		return present("late")
	})
	registry := map[Fact]collector{"slow": blocking}

	start := time.Now()
	collect(context.Background(), Source{}, []Fact{"slow"},
		Options{Timeouts: map[Fact]time.Duration{"slow": 20 * time.Millisecond}}, registry)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("Collect took %v with a 20ms limit", elapsed)
	}
	if got.State != Unknown {
		t.Errorf("fact = %+v, want unknown", got)
	}
}

func TestCollectLateResultDiscarded(t *testing.T) {
	// The collector honors its context but returns a value anyway; the
	// value arrived after the deadline and must not be used.
	var got Value[string]
	late := into(&got, func(ctx context.Context, _ *session) Value[string] {
		<-ctx.Done()
		return present("late")
	})
	collect(context.Background(), Source{}, []Fact{"late"},
		Options{Timeout: 10 * time.Millisecond}, map[Fact]collector{"late": late})
	if got.State != Unknown {
		t.Errorf("fact = %+v, want unknown", got)
	}
}

func TestCollectPanic(t *testing.T) {
	var calls atomic.Int32
	var got Value[string]
	registry := map[Fact]collector{
		"boom": func(context.Context, *session) func(*Info) { panic("parser bug") },
		"ok":   constant(&got, "fine", &calls),
	}
	collect(context.Background(), Source{}, []Fact{"boom", "ok"}, Options{}, registry)
	if got != present("fine") {
		t.Errorf("fact = %+v, want the other collector intact", got)
	}
}

func TestCollectCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got Value[string]
	blocking := into(&got, func(ctx context.Context, _ *session) Value[string] {
		<-ctx.Done()
		return Value[string]{}
	})
	start := time.Now()
	collect(ctx, Source{}, []Fact{"x"}, Options{Timeout: time.Hour}, map[Fact]collector{"x": blocking})
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("Collect took %v with a canceled context", elapsed)
	}
}

func TestOptionsTimeout(t *testing.T) {
	o := Options{Timeout: time.Second, Timeouts: map[Fact]time.Duration{"a": time.Minute, "b": -1}}
	if got := o.timeout("a"); got != time.Minute {
		t.Errorf("override = %v", got)
	}
	if got := o.timeout("b"); got != time.Second {
		t.Errorf("negative override = %v, want the default of the run", got)
	}
	if got := (Options{}).timeout("c"); got != DefaultTimeout {
		t.Errorf("zero options = %v, want %v", got, DefaultTimeout)
	}
}

func TestCollectConcurrentRuns(t *testing.T) {
	var calls atomic.Int32
	registry := map[Fact]collector{"a": func(context.Context, *session) func(*Info) {
		calls.Add(1)
		return func(*Info) {}
	}}
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collect(context.Background(), Source{}, []Fact{"a"}, Options{}, registry)
		}()
	}
	wg.Wait()
	if calls.Load() != 50 {
		t.Errorf("calls = %d, want 50", calls.Load())
	}
}

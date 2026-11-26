package sysinfo

import (
	"context"
	"maps"
	"slices"
	"time"
)

// DefaultTimeout is the time limit of a fact when Options sets none.
const DefaultTimeout = 200 * time.Millisecond

// Options controls a collection run. The zero value uses the defaults.
type Options struct {
	// Timeout is the time limit of every fact without an override. Zero
	// or negative means DefaultTimeout.
	Timeout time.Duration
	// Timeouts overrides Timeout for single facts. Zero or negative
	// values are ignored.
	Timeouts map[Fact]time.Duration
}

// timeout returns the time limit of f.
func (o Options) timeout(f Fact) time.Duration {
	if d := o.Timeouts[f]; d > 0 {
		return d
	}
	if o.Timeout > 0 {
		return o.Timeout
	}
	return DefaultTimeout
}

// collector gathers one fact. It returns a function that stores the
// result in an Info, so that only the goroutine of Collect writes to the
// Info and a collector that finishes late cannot race with the caller.
// A collector must honor ctx and stop the commands it started when ctx
// is done.
type collector func(ctx context.Context, s *session) func(*Info)

// fact adapts a function that returns a typed value to a collector that
// stores it in the field selected by field.
func fact[T any](get func(ctx context.Context, s *session) Value[T], field func(*Info) *Value[T]) collector {
	return func(ctx context.Context, s *session) func(*Info) {
		v := get(ctx, s)
		return func(info *Info) { *field(info) = v }
	}
}

// session is the state shared by the collectors of one Collect call.
// Collectors that share an expensive read keep it here behind a
// sync.Once.
type session struct {
	src  Source
	opts Options
}

// collectors maps every fact sysinfo knows to its collector.
var collectors = map[Fact]collector{}

// Facts returns every fact sysinfo knows, sorted.
func Facts() []Fact {
	return slices.Sorted(maps.Keys(collectors))
}

// Collect gathers the given facts from src concurrently and returns them.
// It never fails: a fact that cannot be determined is Unknown or Absent.
//
// Each fact runs in its own goroutine with its own deadline from opts. A
// fact whose collector does not return in time, or panics, is Unknown,
// and Collect does not wait for it: it returns as soon as every fact is
// done or expired, or ctx is done. Facts that are not requested are not
// collected and stay Unknown. Duplicates are collected once and unknown
// identifiers are ignored.
func Collect(ctx context.Context, src Source, facts []Fact, opts Options) Info {
	return collect(ctx, src, facts, opts, collectors)
}

// collect is Collect with the registry of collectors as a parameter, so
// tests can supply fake collectors.
func collect(ctx context.Context, src Source, facts []Fact, opts Options, registry map[Fact]collector) Info {
	var info Info
	collectLegacy(&info)

	s := &session{src: src, opts: opts}
	seen := make(map[Fact]bool, len(facts))
	results := make(chan func(*Info), len(facts))
	var pending int
	var last time.Time
	for _, f := range facts {
		c, ok := registry[f]
		if !ok || seen[f] {
			continue
		}
		seen[f] = true
		pending++
		fctx, cancel := context.WithTimeout(ctx, opts.timeout(f))
		if deadline, _ := fctx.Deadline(); deadline.After(last) {
			last = deadline
		}
		go func() {
			defer cancel()
			// The channel has room for every result, so a collector
			// that finishes after Collect returned never blocks.
			results <- runCollector(fctx, c, s)
		}()
	}
	if pending == 0 {
		return info
	}

	timer := time.NewTimer(time.Until(last))
	defer timer.Stop()
	for ; pending > 0; pending-- {
		select {
		case apply := <-results:
			if apply != nil {
				apply(&info)
			}
		case <-timer.C:
			return info
		case <-ctx.Done():
			return info
		}
	}
	return info
}

// runCollector runs c and returns its result, or nil when c panicked or
// finished after its deadline. Recovering here is deliberate: a bug in
// one parser must not take down the shell startup.
func runCollector(ctx context.Context, c collector, s *session) (apply func(*Info)) {
	defer func() {
		if recover() != nil {
			apply = nil
		}
	}()
	apply = c(ctx, s)
	if ctx.Err() != nil {
		return nil
	}
	return apply
}

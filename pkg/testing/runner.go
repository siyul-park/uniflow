package testing

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/siyul-park/uniflow/pkg/process"
)

// Runner executes test suites and reports results.
type Runner struct {
	reporters Reporters
	suites    map[string]Suite
	mu        sync.RWMutex
}

// NewRunner creates a new Runner instance.
func NewRunner() *Runner {
	return &Runner{suites: make(map[string]Suite)}
}

// AddReporter adds a reporter if it's not already present.
func (r *Runner) AddReporter(reporter Reporter) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rp := range r.reporters {
		if rp == reporter {
			return false
		}
	}
	r.reporters = append(r.reporters, reporter)
	return true
}

// RemoveReporter removes a reporter if it exists.
func (r *Runner) RemoveReporter(reporter Reporter) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rp := range r.reporters {
		if rp == reporter {
			r.reporters = append(r.reporters[:i], r.reporters[i+1:]...)
			return true
		}
	}
	return false
}

// Register adds a test suite if it's not already registered.
func (r *Runner) Register(name string, suite Suite) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.suites[name]; exists {
		return false
	}
	r.suites[name] = suite
	return true
}

// Unregister removes a registered test suite.
func (r *Runner) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.suites[name]; !exists {
		return false
	}
	delete(r.suites, name)
	return true
}

// Run executes all test suites matching the filter concurrently.
func (r *Runner) Run(ctx context.Context, match func(string) bool) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if match == nil {
		match = func(string) bool { return true }
	}

	g, ctx := errgroup.WithContext(ctx)
	for name, suite := range r.suites {
		if match(name) {
			g.Go(func() error {
				tester := NewTester(name)
				defer tester.Close(nil)

				errors := make(chan error)
				defer close(errors)

				tester.AddExitHook(process.ExitFunc(func(err error) {
					errors <- r.reporters.Report(ctx, &Result{
						ID:        tester.ID(),
						Name:      tester.Name(),
						Error:     err,
						StartTime: tester.StartTime(),
						EndTime:   tester.EndTime(),
					})
				}))

				go func() {
					select {
					case <-ctx.Done():
						tester.Close(ctx.Err())
					case <-tester.Done():
					}
				}()

				go suite.Run(tester)

				return <-errors
			})
		}
	}
	return g.Wait()
}

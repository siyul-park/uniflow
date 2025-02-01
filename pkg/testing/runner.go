package testing

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// Runner executes test suites and reports results.
type Runner struct {
	mu        sync.RWMutex
	reporters Reporters
	suites    map[string]Suite
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
func (r *Runner) Run(ctx context.Context, match func(string) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if match == nil {
		match = func(string) bool { return true }
	}

	var wg sync.WaitGroup
	for name, suite := range r.suites {
		if match(name) {
			wg.Add(1)
			go func() {
				defer wg.Done()

				tester := NewTester(name)
				defer tester.Close(nil)

				go func() {
					select {
					case <-ctx.Done():
						tester.Close(ctx.Err())
					case <-tester.Done():
					}
				}()

				tester.AddExitHook(process.ExitFunc(func(err error) {
					_ = r.reporters.Report(&Result{
						ID:        tester.ID(),
						Name:      tester.Name(),
						Error:     err,
						StartTime: tester.StartTime(),
						EndTime:   tester.EndTime(),
					})
				}))

				suite.Run(tester)
			}()
		}
	}
	wg.Wait()
}

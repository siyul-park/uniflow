package testing

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// Runner represents a test runner that uses a reporter to report test results.
type Runner struct {
	reporters Reporters
	suites    map[string]Suite
	mu        sync.RWMutex
}

// NewRunner creates a new Runner with the provided reporter.
func NewRunner() *Runner {
	return &Runner{suites: make(map[string]Suite)}
}

// AddReporter sets the reporter for the runner.
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

// RemoveReporter removes the specified reporter from the runner.
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

// Register adds a suite to the runner to be executed later.
func (r *Runner) Register(name string, suite Suite) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.suites[name]; exists {
		return false
	}
	r.suites[name] = suite
	return true
}

// Unregister removes a suite from the runner.
func (r *Runner) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.suites[name]; !exists {
		return false
	}
	delete(r.suites, name)
	return true
}

// Run executes all registered test suites that match the given criteria and reports their results.
func (r *Runner) Run(match func(string) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if match == nil {
		match = func(string) bool { return true }
	}

	wg := sync.WaitGroup{}
	for name, s := range r.suites {
		if match(name) {
			wg.Add(1)
			go func() {
				defer wg.Done()

				tester := NewTester(name)
				defer tester.Close(nil)

				tester.Process().AddExitHook(process.ExitFunc(func(err error) {
					_ = r.reporters.Report(&Result{
						ID:        tester.ID(),
						Name:      tester.Name(),
						Error:     err,
						StartTime: tester.Process().StartTime(),
						EndTime:   tester.Process().EndTime(),
					})
				}))

				s.Run(tester)
			}()
		}
	}
	wg.Wait()
}

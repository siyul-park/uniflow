package testing

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// Runner represents a test runner that uses a reporter to report test results.
type Runner struct {
	reporter Reporter
}

// NewRunner creates a new Runner with the provided reporter.
func NewRunner(reporter Reporter) *Runner {
	return &Runner{reporter: reporter}
}

// Run starts a new test and adds an exit hook to report the result.
func (r *Runner) Run(name string) *Tester {
	t := NewTester(name)
	t.Process().AddExitHook(process.ExitFunc(func(err error) {
		_ = r.reporter.Report(&Result{
			Name:      name,
			Error:     err,
			StartTime: t.Process().StartTime(),
			EndTime:   t.Process().EndTime(),
		})
	}))
	return t
}

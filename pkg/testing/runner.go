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
	if reporter == nil {
		reporter = Discard
	}
	return &Runner{reporter: reporter}
}

// Run starts a new test and adds an exit hook to report the result.
func (r *Runner) Run(name string) *Tester {
	t := NewTester(name)
	t.Process().AddExitHook(process.ExitFunc(func(err error) {
		_ = r.reporter.Report(&Result{
			ID:        t.ID(),
			Name:      t.Name(),
			Error:     err,
			StartTime: t.Process().StartTime(),
			EndTime:   t.Process().EndTime(),
		})
	}))
	return t
}

package testing

import (
	"time"
)

// Status represents the current state of a test.
type Status string

const (
	StatusPassed  Status = "PASS"
	StatusFailed  Status = "FAIL"
	StatusSkipped Status = "SKIP"
)

// Result contains information about a test execution.
type Result struct {
	ID        string
	Name      string
	Status    Status
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Children  []*Result
}

// Duration returns the time taken to execute the test.
func (r *Result) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

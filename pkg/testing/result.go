package testing

import "time"

// Result represents the result of a test.
type Result struct {
	Name      string
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// Duration calculates the duration of the test.
func (r *Result) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

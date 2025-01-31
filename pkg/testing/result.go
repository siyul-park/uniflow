package testing

import (
	"time"

	"github.com/gofrs/uuid"
)

// Result represents the result of a test.
type Result struct {
	ID        uuid.UUID
	Name      string
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// Status returns the status of the test as a string.
func (r *Result) Status() string {
	if r.Error != nil {
		return "FAIL"
	}
	return "PASS"
}

// Duration calculates the duration of the test.
func (r *Result) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

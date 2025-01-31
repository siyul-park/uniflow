package testing

import (
	"fmt"
	"io"
)

// Reporter interface defines a method to report test results.
type Reporter interface {
	Report(result *Result) error
}

// ReportFunc is a function type that implements the Reporter interface.
type ReportFunc func(result *Result) error

// Report calls the ReportFunc with the given result.
func (r ReportFunc) Report(result *Result) error {
	return r(result)
}

// NewTextReporter creates a new TextReporter.
func NewTextReporter(logOutput io.Writer) Reporter {
	if logOutput == nil {
		logOutput = io.Discard
	}
	return ReportFunc(func(result *Result) error {
		msg := "passed"
		if result.Error != nil {
			msg = fmt.Sprintf("failed: %v", result.Error)
		}
		_, err := fmt.Fprintf(logOutput, "Test %s %s (Duration: %v)\n", result.Name, msg, result.Duration())
		return err
	})
}

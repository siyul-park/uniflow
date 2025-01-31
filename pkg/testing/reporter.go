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

var Discard = ReportFunc(func(_ *Result) error {
	return nil
})

var _ Reporter = ReportFunc(nil)

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
		if _, err := fmt.Fprintf(logOutput, "%s\t%s\t%v\n", result.Status(), result.Name, result.Duration()); err != nil {
			return err
		}
		if result.Error != nil {
			if _, err := fmt.Fprintf(logOutput, "    %v\n", result.Error); err != nil {
				return err
			}
		}
		return nil
	})
}

package testing

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Reporter interface defines a method to report test results.
type Reporter interface {
	Report(ctx context.Context, result *Result) error
}

// Reporters is a collection of Reporter instances.
type Reporters []Reporter

type reporter struct {
	report func(ctx context.Context, result *Result) error
}

var _ Reporter = (Reporters)(nil)
var _ Reporter = (*reporter)(nil)

// ReportFunc is a function type that implements the Reporter interface.
func ReportFunc(fn func(ctx context.Context, result *Result) error) Reporter {
	return &reporter{report: fn}
}

// NewTextReporter creates a new TextReporter that writes test results to the provided io.Writer.
func NewTextReporter(logOutput io.Writer) Reporter {
	if logOutput == nil {
		logOutput = io.Discard
	}

	var mu sync.Mutex
	return ReportFunc(func(_ context.Context, result *Result) error {
		mu.Lock()
		defer mu.Unlock()

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

// Report reports the test result using all Reporter instances in the collection.
func (r Reporters) Report(ctx context.Context, result *Result) error {
	for _, r := range r {
		if err := r.Report(ctx, result); err != nil {
			return err
		}
	}
	return nil
}

// Report method calls the underlying report function to report the test result.
func (r *reporter) Report(ctx context.Context, result *Result) error {
	return r.report(ctx, result)
}

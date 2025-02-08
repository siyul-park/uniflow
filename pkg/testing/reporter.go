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

// BaseReporter provides common functionality for reporters
type BaseReporter struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewBaseReporter creates a new BaseReporter instance
func NewBaseReporter(w io.Writer) *BaseReporter {
	if w == nil {
		w = io.Discard
	}
	return &BaseReporter{
		writer: w,
	}
}

// TextReporter implements Reporter with basic text output
type TextReporter struct {
	*BaseReporter
}

// NewTextReporter creates a new TextReporter that writes test results to the provided io.Writer
func NewTextReporter(w io.Writer) Reporter {
	return &TextReporter{
		BaseReporter: NewBaseReporter(w),
	}
}

// Report formats and writes test results in a simple text format
func (r *TextReporter) Report(ctx context.Context, result *Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := fmt.Fprintf(r.writer, "%s\t%s\t%v\n", result.Status, result.Name, result.Duration()); err != nil {
		return err
	}
	if result.Error != nil {
		if _, err := fmt.Fprintf(r.writer, "    %v\n", result.Error); err != nil {
			return err
		}
	}
	return nil
}

// PrettyReporter implements Reporter with detailed, formatted output
type PrettyReporter struct {
	*BaseReporter
}

// NewPrettyReporter creates a new PrettyReporter instance
func NewPrettyReporter(w io.Writer) Reporter {
	return &PrettyReporter{
		BaseReporter: NewBaseReporter(w),
	}
}

// Report formats and writes test results with detailed formatting and symbols
func (r *PrettyReporter) Report(ctx context.Context, result *Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Set status and print result
	if result.Error != nil {
		fmt.Fprintf(r.writer, "✗ %s\n", result.Name)
		fmt.Fprintf(r.writer, "  Error: %s\n", result.Error)
		result.Status = StatusFailed
	} else {
		fmt.Fprintf(r.writer, "✓ %s\n", result.Name)
		result.Status = StatusPassed
	}

	// Print duration
	fmt.Fprintf(r.writer, "  Duration: %s\n", result.Duration())
	fmt.Fprintln(r.writer)

	return nil
}

// Report reports the test result using all Reporter instances in the collection
func (r Reporters) Report(ctx context.Context, result *Result) error {
	for _, r := range r {
		if err := r.Report(ctx, result); err != nil {
			return err
		}
	}
	return nil
}

// ReportFunc is a function type that implements the Reporter interface
type ReportFunc func(ctx context.Context, result *Result) error

// Report implements the Reporter interface for ReportFunc
func (f ReportFunc) Report(ctx context.Context, result *Result) error {
	return f(ctx, result)
}

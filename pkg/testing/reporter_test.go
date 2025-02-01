package testing

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReporters_Report(t *testing.T) {
	var reporters Reporters
	reporters = append(reporters, ReportFunc(func(result *Result) error {
		return nil
	}))
	result := &Result{Name: "foo", StartTime: time.Now(), EndTime: time.Now()}
	err := reporters.Report(result)
	assert.NoError(t, err)
}

func TestTextReporter_Report(t *testing.T) {
	t.Run(StatusPass, func(t *testing.T) {
		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", StartTime: time.Now(), EndTime: time.Now()}
		err := reporter.Report(result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "PASS\tfoo")
	})

	t.Run(StatusFail, func(t *testing.T) {
		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", Error: fmt.Errorf("error"), StartTime: time.Now(), EndTime: time.Now()}
		err := reporter.Report(result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "FAIL\tfoo")
	})
}

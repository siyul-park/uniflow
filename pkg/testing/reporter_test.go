package testing

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTextReporter_Report(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", StartTime: time.Now(), EndTime: time.Now()}
		err := reporter.Report(result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "PASS\tfoo")
	})

	t.Run("failed", func(t *testing.T) {
		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", Error: fmt.Errorf("error"), StartTime: time.Now(), EndTime: time.Now()}
		err := reporter.Report(result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "FAIL\tfoo")
	})
}

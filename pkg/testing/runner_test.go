package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Run(t *testing.T) {
	count := 0
	runner := NewRunner(ReportFunc(func(result *Result) error {
		count += 1
		return nil
	}))

	tester := runner.Run("foo")
	assert.NotNil(t, tester)

	tester.Close(nil)
	assert.Equal(t, 1, count)
}

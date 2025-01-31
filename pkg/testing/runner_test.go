package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Run(t *testing.T) {
	runner := NewRunner(NewTextReporter(nil))

	tester := runner.Run("foo")
	assert.NotNil(t, tester)

	tester.Close(nil)
}

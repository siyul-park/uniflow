package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTester_Name(t *testing.T) {
	tester := NewTester("foo")
	assert.Equal(t, "foo", tester.Name())
}

func TestTester_Process(t *testing.T) {
	tester := NewTester("foo")
	assert.NotNil(t, tester.Process())
}

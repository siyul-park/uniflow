package testing

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Register(t *testing.T) {
	runner := NewRunner(nil)
	s := RunFunc(func(tester *Tester) {})

	ok := runner.Register("foo", s)
	assert.True(t, ok)

	ok = runner.Register("foo", s)
	assert.False(t, ok)
}

func TestRunner_Unregister(t *testing.T) {
	runner := NewRunner(nil)
	s := RunFunc(func(tester *Tester) {})

	runner.Register("foo", s)

	ok := runner.Unregister("foo")
	assert.True(t, ok)

	ok = runner.Unregister("foo")
	assert.False(t, ok)
}

func TestRunner_Run(t *testing.T) {
	var count atomic.Int32
	runner := NewRunner(ReportFunc(func(result *Result) error {
		count.Add(1)
		return nil
	}))

	runner.Register("foo", RunFunc(func(tester *Tester) {}))
	runner.Register("bar", RunFunc(func(tester *Tester) {}))

	runner.Run(nil)
	assert.Equal(t, int32(2), count.Load())
}

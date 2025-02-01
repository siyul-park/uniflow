package testing

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunner_AddReporter(t *testing.T) {
	runner := NewRunner()
	r := ReportFunc(func(result *Result) error {
		return nil
	})

	ok := runner.AddReporter(r)
	assert.True(t, ok)

	ok = runner.AddReporter(r)
	assert.False(t, ok)
}

func TestRunner_RemoveReporter(t *testing.T) {
	runner := NewRunner()
	r := ReportFunc(func(result *Result) error {
		return nil
	})

	runner.AddReporter(r)

	ok := runner.RemoveReporter(r)
	assert.True(t, ok)

	ok = runner.RemoveReporter(r)
	assert.False(t, ok)
}

func TestRunner_Register(t *testing.T) {
	runner := NewRunner()
	s := RunFunc(func(tester *Tester) {})

	ok := runner.Register("foo", s)
	assert.True(t, ok)

	ok = runner.Register("foo", s)
	assert.False(t, ok)
}

func TestRunner_Unregister(t *testing.T) {
	runner := NewRunner()
	s := RunFunc(func(tester *Tester) {})

	runner.Register("foo", s)

	ok := runner.Unregister("foo")
	assert.True(t, ok)

	ok = runner.Unregister("foo")
	assert.False(t, ok)
}

func TestRunner_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	runner := NewRunner()

	var count atomic.Int32
	runner.AddReporter(ReportFunc(func(result *Result) error {
		count.Add(1)
		return nil
	}))

	runner.Register("foo", RunFunc(func(tester *Tester) {}))
	runner.Register("bar", RunFunc(func(tester *Tester) {}))

	runner.Run(ctx, nil)
	assert.Equal(t, int32(2), count.Load())
}

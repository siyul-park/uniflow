package testing

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunner_AddReporter(t *testing.T) {
	runner := NewRunner()
	r := ReportFunc(func(_ context.Context, _ *Result) error {
		return nil
	})

	ok := runner.AddReporter(r)
	require.True(t, ok)

	ok = runner.AddReporter(r)
	require.False(t, ok)
}

func TestRunner_RemoveReporter(t *testing.T) {
	runner := NewRunner()
	r := ReportFunc(func(_ context.Context, _ *Result) error {
		return nil
	})

	runner.AddReporter(r)

	ok := runner.RemoveReporter(r)
	require.True(t, ok)

	ok = runner.RemoveReporter(r)
	require.False(t, ok)
}

func TestRunner_Register(t *testing.T) {
	runner := NewRunner()
	s := RunFunc(func(tester *Tester) {})

	ok := runner.Register("foo", s)
	require.True(t, ok)

	ok = runner.Register("foo", s)
	require.False(t, ok)
}

func TestRunner_Unregister(t *testing.T) {
	runner := NewRunner()
	s := RunFunc(func(tester *Tester) {})

	runner.Register("foo", s)

	ok := runner.Unregister("foo")
	require.True(t, ok)

	ok = runner.Unregister("foo")
	require.False(t, ok)
}

func TestRunner_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	runner := NewRunner()

	var count atomic.Int32
	runner.AddReporter(ReportFunc(func(_ context.Context, _ *Result) error {
		count.Add(1)
		return nil
	}))

	runner.Register("foo", RunFunc(func(tester *Tester) {}))
	runner.Register("bar", RunFunc(func(tester *Tester) {}))

	err := runner.Run(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, int32(2), count.Load())
}

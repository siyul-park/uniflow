package testing

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/process"

	"github.com/stretchr/testify/require"
)

func TestTester_ID(t *testing.T) {
	tester := NewTester("foo")
	require.NotZero(t, tester.ID())
}

func TestTester_Name(t *testing.T) {
	tester := NewTester("foo")
	require.Equal(t, "foo", tester.Name())
}

func TestTester_StartTime(t *testing.T) {
	tester := NewTester("foo")
	require.NotZero(t, tester.StartTime())
}

func TestTester_EndTime(t *testing.T) {
	tester := NewTester("foo")
	require.Zero(t, tester.EndTime())
}

func TestTester_Process(t *testing.T) {
	tester := NewTester("foo")
	require.NotNil(t, tester.Process())
}

func TestTester_Err(t *testing.T) {
	tester := NewTester("foo")
	assert.NoError(t, tester.Err())
}

func TestTester_Exit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	tester := NewTester("foo")
	tester.Exit(nil)

	select {
	case <-tester.Done():
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func TestTester_AddExitHook(t *testing.T) {
	tester := NewTester("foo")
	hook := process.ExitFunc(func(err error) {})
	require.True(t, tester.AddExitHook(hook))
}

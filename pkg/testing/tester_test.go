package testing

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/process"

	"github.com/stretchr/testify/assert"
)

func TestTester_ID(t *testing.T) {
	tester := NewTester("foo")
	assert.NotZero(t, tester.ID())
}

func TestTester_Name(t *testing.T) {
	tester := NewTester("foo")
	assert.Equal(t, "foo", tester.Name())
}

func TestTester_StartTime(t *testing.T) {
	tester := NewTester("foo")
	assert.NotZero(t, tester.StartTime())
}

func TestTester_EndTime(t *testing.T) {
	tester := NewTester("foo")
	assert.Zero(t, tester.EndTime())
}

func TestTester_Process(t *testing.T) {
	tester := NewTester("foo")
	assert.NotNil(t, tester.Process())
}

func TestTester_Exit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	tester := NewTester("foo")
	tester.Exit(nil)

	select {
	case <-tester.Done():
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestTester_AddExitHook(t *testing.T) {
	tester := NewTester("foo")
	hook := process.ExitFunc(func(err error) {})
	assert.True(t, tester.AddExitHook(hook))
}

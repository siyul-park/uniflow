package testing

import (
	"context"
	"testing"
	"time"

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

func TestTester_Process(t *testing.T) {
	tester := NewTester("foo")
	assert.NotNil(t, tester.Process())
}

func TestTester_Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	tester := NewTester("foo")
	tester.Close(nil)

	select {
	case <-tester.Done():
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

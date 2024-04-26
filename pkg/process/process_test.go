package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc)
}

func TestProcess_Stack(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc.Stack())
}

func TestProcess_Heap(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc.Heap())
}

func TestProcess_Context(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc.Context())
}

func TestProcess_Lock(t *testing.T) {
	proc := New()

	proc.Lock()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		proc.Close()
		close(done)
	}()

	proc.Unlock()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestProcess_Close(t *testing.T) {
	proc := New()

	count := 0
	proc.AddCloseHook(CloseHookFunc(func() error {
		count += 1
		return nil
	}))

	select {
	case <-proc.Done():
		assert.Fail(t, "proc.Done() is not empty.")
	default:
	}

	proc.Close()

	select {
	case <-proc.Done():
		assert.Equal(t, 1, count)
	default:
		assert.Fail(t, "proc.Done() is empty.")
	}
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Close()
	}
}

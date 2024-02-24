package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc)
}

func TestProcess_Stack(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Stack())
}

func TestProcess_Heap(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Heap())
}

func TestProcess_Lock(t *testing.T) {
	proc := New()

	proc.Lock()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		proc.Exit(nil)
		close(done)
	}()

	proc.Unlock()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestProcess_Exit(t *testing.T) {
	proc := New()

	select {
	case <-proc.Done():
		assert.Fail(t, "proc.Done() is not empty.")
	default:
	}

	proc.Exit(nil)

	select {
	case <-proc.Done():
	default:
		assert.Fail(t, "proc.Done() is empty.")
	}
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Exit(nil)
	}
}

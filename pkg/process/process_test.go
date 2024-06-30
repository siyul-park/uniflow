package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewProcess(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Data())
	assert.NotNil(t, proc.Context())
	assert.Equal(t, nil, proc.Err())
	assert.Equal(t, StatusRunning, proc.Status())
}

func TestProcess_Exit(t *testing.T) {
	proc := New()

	proc.Exit(nil)
	assert.Equal(t, nil, proc.Err())
	assert.Equal(t, StatusTerminated, proc.Status())
}

func TestProcess_AtExit(t *testing.T) {
	proc := New()

	count := 0
	h := ExitHookFunc(func(err error) {
		count++
	})
	proc.AtExit(h)

	proc.Exit(nil)
	assert.Equal(t, 1, count)
}

func TestProcess_Fork(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	child := proc.Fork()
	defer child.Exit(nil)

	assert.Equal(t, proc, child.Parent())
}

func TestProcess_Wait(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	child := proc.Fork()
	defer child.Exit(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		child.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	done = make(chan struct{})
	go func() {
		proc.Wait()
		close(done)
	}()

	child.Exit(nil)

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Exit(nil)
	}
}

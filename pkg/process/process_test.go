package process

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc)
}

func TestProcess_ID(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotEqual(t, uuid.UUID{}, proc.ID())
}

func TestProcess_Graph(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Graph())
}

func TestProcess_Stack(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Stack())
}

func TestProcess_Share(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Share())
}

func TestProcess_Lock(t *testing.T) {
	proc := New()

	proc.Lock()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		proc.Stack().Done(uuid.UUID{})
		close(done)
	}()

	proc.Unlock()

	select {
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	case <-done:
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

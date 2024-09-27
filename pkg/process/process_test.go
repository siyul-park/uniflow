package process

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewProcess(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotZero(t, proc.ID())
	assert.NotNil(t, proc.Context())
	assert.Equal(t, nil, proc.Err())
	assert.Equal(t, StatusRunning, proc.Status())
}

func TestProcess_Load(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	r := proc.Load(k)
	assert.Nil(t, r)

	proc.Store(k, v)

	r = proc.Load(k)
	assert.Equal(t, v, r)
}

func TestProcess_Store(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	proc.Store(k, v1)
	proc.Store(k, v2)

	r := proc.Load(k)
	assert.Equal(t, v2, r)
}

func TestProcess_Delete(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	ok := proc.Delete(k)
	assert.False(t, ok)

	proc.Store(k, v)

	ok = proc.Delete(k)
	assert.True(t, ok)
}

func TestProcess_LoadAndDelete(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	proc.Store(k, v)

	r := proc.LoadAndDelete(k)
	assert.Equal(t, v, r)

	r = proc.Load(k)
	assert.Nil(t, r)
}

func TestProcess_Exit(t *testing.T) {
	proc := New()

	proc.Exit(nil)
	assert.Equal(t, nil, proc.Err())
	assert.Equal(t, StatusTerminated, proc.Status())
}

func TestProcess_AddExitHook(t *testing.T) {
	proc := New()

	count := 0
	h := ExitFunc(func(err error) {
		count++
	})
	proc.AddExitHook(h)

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

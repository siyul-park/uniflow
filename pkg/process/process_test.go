package process

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotZero(t, proc.ID())
	assert.Equal(t, nil, proc.Err())
	assert.NotZero(t, proc.StartTime())
	assert.Equal(t, StatusRunning, proc.Status())
}

func TestProcess_Keys(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	proc.SetValue(k, v)

	keys := proc.Keys()
	assert.Contains(t, keys, k)
}

func TestProcess_Value(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	r := proc.Value(k)
	assert.Nil(t, r)

	proc.SetValue(k, v)

	r = proc.Value(k)
	assert.Equal(t, v, r)
}

func TestProcess_SetValue(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	proc.SetValue(k, v1)
	proc.SetValue(k, v2)

	r := proc.Value(k)
	assert.Equal(t, v2, r)
}

func TestProcess_RemoveValue(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	proc.SetValue(k, v)

	r := proc.RemoveValue(k)
	assert.Equal(t, v, r)

	r = proc.Value(k)
	assert.Nil(t, r)
}

func TestProcess_Exit(t *testing.T) {
	proc := New()

	proc.Exit(nil)
	assert.ErrorIs(t, proc.Err(), context.Canceled)
	assert.NotZero(t, proc.EndTime())
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

func TestProcess_Join(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	child := proc.Fork()
	defer child.Exit(nil)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		child.Join()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	done = make(chan struct{})
	go func() {
		proc.Join()
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

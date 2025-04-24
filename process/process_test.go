package process

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	require.NotZero(t, proc.ID())
	require.Equal(t, nil, proc.Err())
	require.NotZero(t, proc.StartTime())
	require.Equal(t, StatusRunning, proc.Status())
}

func TestProcess_Keys(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	proc.SetValue(k, v)

	keys := proc.Keys()
	require.Contains(t, keys, k)
}

func TestProcess_Value(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	r := proc.Value(k)
	require.Nil(t, r)

	proc.SetValue(k, v)

	r = proc.Value(k)
	require.Equal(t, v, r)
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
	require.Equal(t, v2, r)
}

func TestProcess_RemoveValue(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	proc.SetValue(k, v)

	r := proc.RemoveValue(k)
	require.Equal(t, v, r)

	r = proc.Value(k)
	require.Nil(t, r)
}

func TestProcess_Exit(t *testing.T) {
	proc := New()

	proc.Exit(nil)
	require.ErrorIs(t, proc.Err(), context.Canceled)
	require.NotZero(t, proc.EndTime())
	require.Equal(t, StatusTerminated, proc.Status())
}

func TestProcess_AddExitHook(t *testing.T) {
	proc := New()

	count := 0
	h := ExitFunc(func(err error) {
		count++
	})
	proc.AddExitHook(h)

	proc.Exit(nil)
	require.Equal(t, 1, count)
}

func TestProcess_Fork(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	child := proc.Fork()
	defer child.Exit(nil)

	require.Equal(t, proc, child.Parent())
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
		require.NoError(t, ctx.Err())
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
		require.NoError(t, ctx.Err())
	}
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Exit(nil)
	}
}

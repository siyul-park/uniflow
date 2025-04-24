package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/require"
)

func TestInPort_Open(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	r1 := in.Open(proc)
	r2 := in.Open(proc)

	require.Equal(t, r1, r2)
}

func TestInPort_OpenHook(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	done := make(chan struct{})
	h := OpenHookFunc(func(proc *process.Process) {
		close(done)
	})

	ok := in.AddOpenHook(h)
	require.True(t, ok)

	ok = in.AddOpenHook(h)
	require.False(t, ok)

	_ = in.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}

	ok = in.RemoveOpenHook(h)
	require.True(t, ok)

	ok = in.RemoveOpenHook(h)
	require.False(t, ok)
}

func TestInPort_CloseHook(t *testing.T) {
	in := NewIn()
	defer in.Close()

	done := make(chan struct{})
	h := CloseHookFunc(func() {
		close(done)
	})

	ok := in.AddCloseHook(h)
	require.True(t, ok)

	ok = in.AddCloseHook(h)
	require.False(t, ok)

	in.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
}

func TestInPort_Listener(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	done := make(chan struct{})
	h := ListenFunc(func(proc *process.Process) {
		close(done)
	})

	ok := in.AddListener(h)
	require.True(t, ok)

	ok = in.AddListener(h)
	require.False(t, ok)

	_ = in.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
}

func BenchmarkInPort_Open(b *testing.B) {
	in := NewIn()
	defer in.Close()

	b.RunParallel(func(p *testing.PB) {
		proc := process.New()
		defer proc.Exit(nil)

		for p.Next() {
			in.Open(proc)
		}
	})
}

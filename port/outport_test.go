package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/require"
)

func TestOutPort_Open(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	w1 := out.Open(proc)
	w2 := out.Open(proc)

	require.Equal(t, w1, w2)
}

func TestOutPort_Link(t *testing.T) {
	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	ok := out.Link(in)
	require.True(t, ok)
	require.Len(t, out.Links(), 1)

	ok = out.Link(in)
	require.False(t, ok)
}

func TestOutPort_Unlink(t *testing.T) {
	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	out.Link(in)

	ok := out.Unlink(in)
	require.True(t, ok)
	require.Len(t, out.Links(), 0)

	ok = out.Unlink(in)
	require.False(t, ok)
}

func TestOutPort_OpenHook(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	h := OpenHookFunc(func(proc *process.Process) {
		close(done)
	})

	ok := out.AddOpenHook(h)
	require.True(t, ok)

	ok = out.AddOpenHook(h)
	require.False(t, ok)

	_ = out.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
}

func TestOutPort_CloseHook(t *testing.T) {
	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	h := CloseHookFunc(func() {
		close(done)
	})

	ok := out.AddCloseHook(h)
	require.True(t, ok)

	ok = out.AddCloseHook(h)
	require.False(t, ok)

	out.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
}

func TestOutPort_Listener(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	h := ListenFunc(func(proc *process.Process) {
		close(done)
	})

	ok := out.AddListener(h)
	require.True(t, ok)

	ok = out.AddListener(h)
	require.False(t, ok)

	_ = out.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
}

func BenchmarkOutPort_Open(b *testing.B) {
	out := NewOut()
	defer out.Close()

	b.RunParallel(func(p *testing.PB) {
		proc := process.New()
		defer proc.Exit(nil)

		for p.Next() {
			out.Open(proc)
		}
	})
}

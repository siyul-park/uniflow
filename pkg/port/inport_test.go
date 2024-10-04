package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestInPort_Open(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	r1 := in.Open(proc)
	r2 := in.Open(proc)

	assert.Equal(t, r1, r2)
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
	assert.True(t, ok)

	ok = in.AddOpenHook(h)
	assert.False(t, ok)

	_ = in.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	ok = in.RemoveOpenHook(h)
	assert.True(t, ok)

	ok = in.RemoveOpenHook(h)
	assert.False(t, ok)
}

func TestInPort_CloseHook(t *testing.T) {
	in := NewIn()
	defer in.Close()

	done := make(chan struct{})
	h := CloseHookFunc(func() {
		close(done)
	})

	ok := in.AddCloseHook(h)
	assert.True(t, ok)

	ok = in.AddCloseHook(h)
	assert.False(t, ok)

	in.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
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
	assert.True(t, ok)

	ok = in.AddListener(h)
	assert.False(t, ok)

	_ = in.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
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

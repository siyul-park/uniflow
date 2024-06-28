package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/assert"
)

func TestOutPort_Open(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	w1 := out.Open(proc)
	w2 := out.Open(proc)

	assert.Equal(t, w1, w2)
}

func TestOutPort_Link(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	out.Link(in)
	assert.Equal(t, 1, out.Links())
}

func TestOutPort_AddInitHook(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	h := InitHookFunc(func(proc *process.Process) {
		close(done)
	})

	out.AddInitHook(h)

	_ = out.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
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

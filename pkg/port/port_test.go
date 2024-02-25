package port

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestPort_Open(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	out.Link(in)

	r := in.Open(proc)
	w := out.Open(proc)

	assert.Equal(t, r, in.Open(proc))
	assert.Equal(t, w, out.Open(proc))
}

func TestPort_Link(t *testing.T) {
	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	out.Link(in)

	assert.Equal(t, 1, out.Links())
}

func TestPort_AddHandler(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	count := atomic.Int32{}
	h := HandlerFunc(func(proc *process.Process) {
		if count.Add(1) == 2 {
			close(done)
		}
	})

	in.AddHandler(h)
	out.AddHandler(h)

	_ = in.Open(proc)
	_ = out.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func BenchmarkPort_Open(b *testing.B) {
	in := NewIn()
	defer in.Close()

	out := NewOut()
	defer out.Close()

	out.Link(in)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			proc := process.New()
			defer proc.Close()

			out.Open(proc)
			in.Open(proc)
		}
	})
}

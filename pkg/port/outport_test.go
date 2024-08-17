package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	out := NewOut()
	defer out.Close()

	res, err := Write(out, nil)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

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

func TestOutPort_AddAndRemoveListener(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	out := NewOut()
	defer out.Close()

	done := make(chan struct{})
	h := ListenFunc(func(proc *process.Process) {
		close(done)
	})

	ok := out.AddListener(h)
	assert.True(t, ok)

	ok = out.AddListener(h)
	assert.False(t, ok)

	_ = out.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	ok = out.RemoveListener(h)
	assert.True(t, ok)

	ok = out.RemoveListener(h)
	assert.False(t, ok)
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

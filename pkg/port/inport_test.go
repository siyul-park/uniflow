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

func TestInPort_AddInitHook(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	in := NewIn()
	defer in.Close()

	done := make(chan struct{})
	h := InitHookFunc(func(proc *process.Process) {
		close(done)
	})

	in.AddInitHook(h)

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

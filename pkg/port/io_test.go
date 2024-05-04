package port

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestIO_WriteAndRead(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	w := newWriter(proc, 0)
	defer w.Close()

	r := newReader(proc, 0)
	defer r.Close()

	w.link(r)

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	w.Write(pck1)
	w.Write(pck2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case pck := <-r.Read():
		r.Receive(pck)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case pck := <-r.Read():
		proc.Stack().Clear(pck)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-w.Receive():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestIO_Link(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	w := newWriter(proc, 0)
	defer w.Close()

	r := newReader(proc, 0)
	defer r.Close()

	w.link(r)
	assert.Equal(t, 1, w.Links())
	assert.Equal(t, 1, r.Links())

	w.unlink(r)
	assert.Equal(t, 0, w.Links())
	assert.Equal(t, 0, r.Links())
}

func TestIO_Done(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	w := newWriter(proc, 0)
	r := newReader(proc, 0)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	w.Close()
	r.Close()

	select {
	case <-w.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-r.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestIO_Cost(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	w := newWriter(proc, 0)
	defer w.Close()

	r := newReader(proc, 0)
	defer r.Close()

	w.link(r)

	pck := packet.New(nil)

	w.Write(pck)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case pck := <-r.Read():
		assert.Equal(t, 0, r.Cost(pck))
		r.Receive(pck)
		assert.Equal(t, math.MaxInt, r.Cost(pck))
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-w.Receive():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestDiscard(t *testing.T) {
	proc := process.New()
	defer proc.Close()

	w := newWriter(proc, 0)
	defer w.Close()

	r := newReader(proc, 0)
	defer r.Close()

	w.link(r)

	Discard(w)

	pck := packet.New(nil)

	w.Write(pck)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case pck := <-r.Read():
		r.Receive(pck)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func BenchmarkIO_WriteAndRead(b *testing.B) {
	proc := process.New()
	defer proc.Close()
	defer proc.Stack().Close()

	w := newWriter(proc, 0)
	defer w.Close()

	r := newReader(proc, 0)
	defer r.Close()

	w.link(r)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			outPck := packet.New(nil)
			w.Write(outPck)

			backPck := <-r.Read()
			r.Receive(backPck)

			<-w.Receive()
		}
	})
}

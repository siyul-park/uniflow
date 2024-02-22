package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestPipe_WriteAndRead(t *testing.T) {
	p1 := newPipe(0)
	defer p1.Close()

	p2 := newPipe(0)
	defer p2.Close()

	p1.Link(p2)

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	p1.Write(pck1)
	p1.Write(pck2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case pck := <-p2.Read():
		p2.Write(pck)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case pck := <-p2.Read():
		p2.Write(pck)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-p1.Read():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-p1.Read():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestPipe_Link(t *testing.T) {
	p1 := newPipe(0)
	defer p1.Close()

	p2 := newPipe(0)
	defer p2.Close()

	p1.Link(p2)
	assert.Equal(t, 1, p1.Links())
	assert.Equal(t, 1, p2.Links())

	p1.Unlink(p2)
	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())
}

func TestPipe_Done(t *testing.T) {
	p := newPipe(0)

	p.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-p.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func BenchmarkPipe_WriteAndRead(b *testing.B) {
	p1 := newPipe(0)
	defer p1.Close()

	p2 := newPipe(0)
	defer p2.Close()

	p1.Link(p2)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			p1.Write(packet.New(nil))
			<-p2.Read()
		}
	})
}

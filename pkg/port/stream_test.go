package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestStream_New(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	stream := newStream(proc)

	select {
	case <-stream.Done():
		assert.Fail(t, "stream.Done() is empty.")
	default:
	}
}

func TestStream_Link(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	stream1 := newStream(proc)
	stream2 := newStream(proc)

	stream1.Link(stream2)

	count := 0
	stream1.AddSendHook(SendHookFunc(func(_ *packet.Packet) {
		count += 1
	}))

	pck := packet.New(nil)

	stream1.Send(pck)
	assert.Equal(t, 1, count)
	assert.Equal(t, pck, <-stream2.Receive())
}

func TestStream_Unlink(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	stream1 := newStream(proc)
	stream2 := newStream(proc)

	stream1.Link(stream2)
	stream1.Unlink(stream2)

	pck := packet.New(nil)

	stream1.Send(pck)

	select {
	case <-stream2.Receive():
		assert.Fail(t, "stream should not receive and packet.")
	default:
	}
}

func TestStream_Close(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	stream := newStream(proc)

	stream.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-stream.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func BenchmarkStream_SendAndReceive(b *testing.B) {
	proc := process.New()
	defer proc.Exit(nil)

	in := newStream(proc)
	defer in.Close()
	out := newStream(proc)
	defer in.Close()

	in.Link(out)

	pck := packet.New(nil)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			in.Send(pck)
			<-out.Receive()
		}
	})
}

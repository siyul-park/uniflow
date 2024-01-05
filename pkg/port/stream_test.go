package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestStream_New(t *testing.T) {
	stream := newStream()

	select {
	case <-stream.Done():
		assert.Fail(t, "stream.Done() is empty.")
	default:
	}
}

func TestStream_Link(t *testing.T) {
	stream1 := newStream()
	stream2 := newStream()

	stream1.Link(stream2)

	pck := packet.New(nil)

	stream1.Send(pck)

	assert.Equal(t, pck, <-stream2.Receive())
}

func TestStream_Unlink(t *testing.T) {
	stream1 := newStream()
	stream2 := newStream()

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

func BenchmarkStream_SendAndReceive(b *testing.B) {
	in := newStream()
	defer in.Close()
	out := newStream()
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

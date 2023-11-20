package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestStream_New(t *testing.T) {
	stream := NewStream()

	select {
	case <-stream.Done():
		assert.Fail(t, "stream.Done() is empty.")
	default:
	}
}

func TestStream_Link(t *testing.T) {
	stream1 := NewStream()
	stream2 := NewStream()

	stream1.Link(stream2)

	pck := packet.New(nil)

	stream1.Send(pck)

	assert.Equal(t, pck, <-stream2.Receive())
}

func TestStream_Unlink(t *testing.T) {
	stream1 := NewStream()
	stream2 := NewStream()

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

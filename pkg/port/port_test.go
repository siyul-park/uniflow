package port

import (
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	port := New()
	defer port.Close()

	assert.NotNil(t, port)
}

func TestPort_Link(t *testing.T) {
	port1 := New()
	defer port1.Close()
	port2 := New()
	defer port2.Close()

	port1.Link(port2)

	proc := process.New()

	stream1 := port1.Open(proc)
	stream2 := port2.Open(proc)

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	stream1.Send(pck1)
	stream2.Send(pck2)

	assert.Equal(t, pck1, <-stream2.Receive())
	assert.Equal(t, pck2, <-stream1.Receive())
}

func TestPort_UnLink(t *testing.T) {
	port1 := New()
	defer port1.Close()
	port2 := New()
	defer port2.Close()

	port1.Link(port2)
	port1.Unlink(port2)

	proc := process.New()

	stream1 := port1.Open(proc)
	stream2 := port2.Open(proc)

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	stream1.Send(pck1)
	stream2.Send(pck2)

	select {
	case <-stream1.Receive():
		assert.Fail(t, "pipe should not receive and packet.")
	default:
	}
	select {
	case <-stream2.Receive():
		assert.Fail(t, "pipe should not receive and packet.")
	default:
	}
}

func TestPortLinks(t *testing.T) {
	port1 := New()
	defer port1.Close()
	port2 := New()
	defer port2.Close()

	assert.Equal(t, port1.Links(), 0)
	assert.Equal(t, port2.Links(), 0)

	port1.Link(port2)

	assert.Equal(t, port1.Links(), 1)
	assert.Equal(t, port2.Links(), 1)
}

func TestPort_Open(t *testing.T) {
	port := New()
	defer port.Close()

	t.Run("process not closed", func(t *testing.T) {
		proc := process.New()
		stream := port.Open(proc)

		proc.Exit()

		select {
		case <-stream.Done():
		case <-time.Tick(time.Second):
			assert.Fail(t, "pipe.Done() is empty.")
		}
	})

	t.Run("process closed", func(t *testing.T) {
		proc := process.New()
		proc.Exit()

		stream := port.Open(proc)

		select {
		case <-stream.Done():
		default:
			assert.Fail(t, "stream.Done() is empty.")
		}
	})
}

func TestPort_Close(t *testing.T) {
	port := New()
	defer port.Close()
	proc := process.New()
	stream := port.Open(proc)

	port.Close()

	select {
	case <-stream.Done():
	default:
		assert.Fail(t, "stream.Done() is empty.")
	}

	select {
	case <-port.Done():
	default:
		assert.Fail(t, "port.Done() is empty.")
	}
}

func BenchmarkPort_Open(b *testing.B) {
	port := New()
	defer port.Close()

	for i := 0; i < b.N; i++ {
		proc := process.New()
		_ = port.Open(proc)
	}
}

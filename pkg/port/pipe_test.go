package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestPipe_Link(t *testing.T) {
	t.Run("1:1", func(t *testing.T) {
		read := NewReadPipe()
		defer read.Close()
		write := NewWritePipe()
		defer write.Close()

		write.Link(read)

		pck := packet.New(nil)
		write.Send(pck)

		assert.Equal(t, pck, <-read.Receive())
	})

	t.Run("1:N", func(t *testing.T) {
		read1 := NewReadPipe()
		defer read1.Close()
		read2 := NewReadPipe()
		defer read2.Close()
		write := NewWritePipe()
		defer write.Close()

		write.Link(read1)
		write.Link(read2)

		pck := packet.New(nil)
		write.Send(pck)

		assert.Equal(t, pck, <-read1.Receive())
		assert.Equal(t, pck, <-read2.Receive())
	})
}

func TestPipe_Unlink(t *testing.T) {
	read := NewReadPipe()
	defer read.Close()
	write := NewWritePipe()
	defer write.Close()

	write.Link(read)
	write.Unlink(read)

	pck := packet.New(nil)
	write.Send(pck)

	select {
	case <-read.Receive():
		assert.Fail(t, "pipe should not receive and packet.")
	default:
	}
}

func TestPipe_SendAndReceive(t *testing.T) {
	t.Run("Not Closed", func(t *testing.T) {
		read := NewReadPipe()
		defer read.Close()
		write := NewWritePipe()
		defer write.Close()

		write.Link(read)

		pck1 := packet.New(nil)
		pck2 := packet.New(nil)

		write.Send(pck1)
		write.Send(pck2)

		assert.Equal(t, pck1, <-read.Receive())
		assert.Equal(t, pck2, <-read.Receive())
	})

	t.Run("Closed", func(t *testing.T) {
		read := NewReadPipe()
		defer read.Close()
		write := NewWritePipe()
		defer write.Close()

		write.Link(read)
		write.Close()

		pck1 := packet.New(nil)
		pck2 := packet.New(nil)

		write.Send(pck1)
		write.Send(pck2)

		assert.Nil(t, <-read.Receive())
		assert.Nil(t, <-read.Receive())
	})
}

func TestPipe_Close(t *testing.T) {
	t.Run("ReadPipe", func(t *testing.T) {
		pipe := NewReadPipe()
		defer pipe.Close()

		select {
		case <-pipe.Done():
			assert.Fail(t, "pipe.Done() is not empty.")
		default:
		}

		pipe.Close()

		select {
		case <-pipe.Done():
		default:
			assert.Fail(t, "pipe.Done() is empty.")
		}
	})
	t.Run("WritePipe", func(t *testing.T) {
		pipe := NewWritePipe()
		defer pipe.Close()

		select {
		case <-pipe.Done():
			assert.Fail(t, "pipe.Done() is not empty.")
		default:
		}

		pipe.Close()

		select {
		case <-pipe.Done():
		default:
			assert.Fail(t, "pipe.Done() is empty.")
		}
	})
}

func BenchmarkPipe_SendAndReceive(b *testing.B) {
	read := NewReadPipe()
	defer read.Close()
	write := NewWritePipe()
	defer write.Close()

	write.Link(read)

	pck := packet.New(nil)

	for i := 0; i < b.N; i++ {
		write.Send(pck)
		<-read.Receive()
	}
}

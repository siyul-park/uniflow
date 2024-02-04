package port

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestPipe_Link(t *testing.T) {
	t.Run("1:1", func(t *testing.T) {
		proc := process.New()
		defer proc.Exit(nil)

		read := newReadPipe(proc)
		defer read.Close()
		write := newWritePipe(proc)
		defer write.Close()

		write.Link(read)

		pck := packet.New(nil)
		write.Send(pck)

		assert.Equal(t, pck, <-read.Receive())
	})

	t.Run("1:N", func(t *testing.T) {
		proc := process.New()
		defer proc.Exit(nil)

		read1 := newReadPipe(proc)
		defer read1.Close()
		read2 := newReadPipe(proc)
		defer read2.Close()
		write := newWritePipe(proc)
		defer write.Close()

		write.Link(read1)
		write.Link(read2)

		assert.Equal(t, 2, write.Links())

		pck := packet.New(nil)
		write.Send(pck)

		assert.True(t, proc.Graph().Has(pck.ID(), (<-read1.Receive()).ID()))
		assert.True(t, proc.Graph().Has(pck.ID(), (<-read2.Receive()).ID()))
	})
}

func TestPipe_Unlink(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	read := newReadPipe(proc)
	defer read.Close()
	write := newWritePipe(proc)
	defer write.Close()

	write.Link(read)
	write.Unlink(read)

	assert.Equal(t, 0, write.Links())

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
		proc := process.New()
		defer proc.Exit(nil)

		read := newReadPipe(proc)
		defer read.Close()
		write := newWritePipe(proc)
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
		proc := process.New()
		defer proc.Exit(nil)

		read := newReadPipe(proc)
		defer read.Close()
		write := newWritePipe(proc)
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
		proc := process.New()
		defer proc.Exit(nil)

		pipe := newReadPipe(proc)
		defer pipe.Close()

		pck := packet.New(nil)
		pipe.send(pck)

		go pipe.Close()

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case <-pipe.Done():
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}
	})

	t.Run("WritePipe", func(t *testing.T) {
		proc := process.New()
		defer proc.Exit(nil)

		pipe := newWritePipe(proc)
		defer pipe.Close()

		pipe.Close()

		select {
		case <-pipe.Done():
		default:
			assert.Fail(t, "pipe.Done() is empty.")
		}
	})
}

func TestPipe_SendHook(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	read1 := newReadPipe(proc)
	defer read1.Close()
	read2 := newReadPipe(proc)
	defer read2.Close()

	write := newWritePipe(proc)
	defer write.Close()

	write.Link(read1)
	write.Link(read2)

	pck := packet.New(nil)

	count := 0
	write.AddSendHook(SendHookFunc(func(_ *packet.Packet) {
		count += 1
	}))

	write.Send(pck)
	assert.Equal(t, 2, count)
}

func BenchmarkPipe_SendAndReceive(b *testing.B) {
	proc := process.New()
	defer proc.Exit(nil)

	read := newReadPipe(proc)
	defer read.Close()
	write := newWritePipe(proc)
	defer write.Close()

	write.Link(read)

	pck := packet.New(nil)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			write.Send(pck)
			<-read.Receive()
		}
	})
}

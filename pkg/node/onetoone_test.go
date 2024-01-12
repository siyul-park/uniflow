package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewOneToOneNode(t *testing.T) {
	n := NewOneToOneNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestOneToOneNode_Port(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p, ok := n.Port(PortIO)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestOneToOneNode_SendAndReceive(t *testing.T) {
	t.Run("IO", func(t *testing.T) {
		t.Run("With Out Port", func(t *testing.T) {
			n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return inPck, nil
			})
			defer n.Close()

			io := port.New()
			ioPort, _ := n.Port(PortIO)
			ioPort.Link(io)

			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-ioStream.Receive():
				assert.Equal(t, inPayload, outPck.Payload())
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})

		t.Run("With Err Port", func(t *testing.T) {
			n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
			})
			defer n.Close()

			io := port.New()
			ioPort, _ := n.Port(PortIO)
			ioPort.Link(io)

			err := port.New()
			errPort, _ := n.Port(PortErr)
			errPort.Link(err)

			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)
			errStream := err.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-errStream.Receive():
				assert.NotNil(t, outPck)
				errStream.Send(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}

			select {
			case backPck := <-ioStream.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})
	})

	t.Run("In/Out", func(t *testing.T) {
		t.Run("With Out Port", func(t *testing.T) {
			n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return inPck, nil
			})
			defer n.Close()

			in := port.New()
			inPort, _ := n.Port(PortIn)
			inPort.Link(in)

			out := port.New()
			outPort, _ := n.Port(PortOut)
			outPort.Link(out)

			proc := process.New()
			defer proc.Exit(nil)

			inStream := in.Open(proc)
			outStream := out.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			inStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-outStream.Receive():
				assert.Equal(t, inPayload, outPck.Payload())
				outStream.Send(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}

			select {
			case backPck := <-inStream.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})

		t.Run("With Err Port", func(t *testing.T) {
			n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
			})
			defer n.Close()

			in := port.New()
			inPort, _ := n.Port(PortIn)
			inPort.Link(in)

			err := port.New()
			errPort, _ := n.Port(PortErr)
			errPort.Link(err)

			proc := process.New()
			defer proc.Exit(nil)

			inStream := in.Open(proc)
			errStream := err.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			inStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-errStream.Receive():
				assert.NotNil(t, outPck)
				errStream.Send(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}

			select {
			case backPck := <-inStream.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})
	})
}

func BenchmarkOneToOneNode_SendAndReceive(b *testing.B) {
	b.Run("IO", func(b *testing.B) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioStream.Send(inPck)
			<-ioStream.Receive()
		}
	})

	b.Run("In/Out", func(b *testing.B) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		in := port.New()
		inPort, _ := n.Port(PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(PortOut)
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				inStream.Send(inPck)
				<-outStream.Receive()
			}
		})
	})
}

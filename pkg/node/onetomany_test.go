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

func TestNewOneToManyNode(t *testing.T) {
	n := NewOneToManyNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestOneToManyNode_Port(t *testing.T) {
	n := NewOneToManyNode(nil)
	defer func() { _ = n.Close() }()

	p, ok := n.Port(PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(port.SetIndex(PortOut, 0))
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestOneToManyNode_Send(t *testing.T) {
	t.Run("return out", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck}, nil
		})
		defer func() { _ = n.Close() }()

		in := port.New()
		inPort, _ := n.Port(PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(port.SetIndex(PortOut, 0))
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
			select {
			case outPck := <-inStream.Receive():
				assert.NotNil(t, outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("return err", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, packet.New(primitive.NewString(faker.Word()))
		})
		defer func() { _ = n.Close() }()

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
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkOneToManyNode_Send(b *testing.B) {
	n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
		return []*packet.Packet{inPck}, nil
	})
	defer func() { _ = n.Close() }()

	in := port.New()
	inPort, _ := n.Port(PortIn)
	inPort.Link(in)

	out := port.New()
	outPort, _ := n.Port(port.SetIndex(PortOut, 0))
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inStream := in.Open(proc)
	outStream := out.Open(proc)

	inPayload := primitive.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inStream.Send(inPck)
		<-outStream.Receive()
	}
}

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
	defer n.Close()

	p := n.Port(PortIn)
	assert.NotNil(t, p)

	p = n.Port(MultiPort(PortOut, 0))
	assert.NotNil(t, p)

	p = n.Port(PortErr)
	assert.NotNil(t, p)
}

func TestOneToManyNode_SendAndReceive(t *testing.T) {
	t.Run("With Out Port", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck}, nil
		})
		defer n.Close()

		in := port.New()
		inPort := n.Port(PortIn)
		inPort.Link(in)

		out := port.New()
		outPort := n.Port(MultiPort(PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			outStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("With Err Port", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
		})
		defer n.Close()

		in := port.New()
		inPort := n.Port(PortIn)
		inPort.Link(in)

		err := port.New()
		errPort := n.Port(PortErr)
		errPort.Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		errStream := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errStream.Receive():
			assert.NotNil(t, outPck)
			errStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkOneToManyNode_SendAndReceive(b *testing.B) {
	n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
		return []*packet.Packet{inPck}, nil
	})
	defer n.Close()

	in := port.New()
	inPort := n.Port(PortIn)
	inPort.Link(in)

	out := port.New()
	outPort := n.Port(MultiPort(PortOut, 0))
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)
	defer proc.Stack().Close()

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
}

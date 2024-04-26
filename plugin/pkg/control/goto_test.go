package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewGoToNode(t *testing.T) {
	n := NewGoToNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestGoToNode_Port(t *testing.T) {
	n := NewGoToNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestGoToNode_SendAndReceive(t *testing.T) {
	t.Run("In -> Out0 -> Out1", func(t *testing.T) {
		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n1.Close()

		n2 := NewGoToNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIO))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		out1 := port.NewIn()
		n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader1.Read():
			assert.Equal(t, inPayload, outPck.Payload())
			outReader1.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Out0 -> Error -> In", func(t *testing.T) {
		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.WithError(errors.New(faker.Sentence()), inPck)
		})
		defer n1.Close()

		n2 := NewGoToNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIO))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		err := port.NewIn()
		n2.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errReader.Read():
			assert.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func TestGoToNodeCodec_Decode(t *testing.T) {
	codec := NewGoToNodeCodec()

	spec := &GoToNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func BenchmarkGoToNode_SendAndReceive(b *testing.B) {
	n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})
	defer n1.Close()

	n2 := NewGoToNode()
	defer n2.Close()

	n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIO))

	in := port.NewOut()
	in.Link(n2.In(node.PortIn))

	out1 := port.NewIn()
	n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

	proc := process.New()
	defer proc.Close()

	inWriter := in.Open(proc)
	outReader1 := out1.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		outPck := <-outReader1.Read()
		outReader1.Receive(outPck)

		<-inWriter.Receive()
	}
}

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

func TestNewJumpNode(t *testing.T) {
	n := NewJumpNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestJumpNode_Port(t *testing.T) {
	n := NewJumpNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortIO))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestJumpNode_SendAndReceive(t *testing.T) {
	t.Run("In -> IO -> Out", func(t *testing.T) {
		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n1.Close()

		n2 := NewJumpNode()
		defer n2.Close()

		n2.Out(node.PortIO).Link(n1.In(node.PortIO))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		out := port.NewIn()
		n2.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, inPayload, outPck.Payload())
			outReader.Receive(outPck)
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

	t.Run("In -> IO -> Error -> In", func(t *testing.T) {
		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.WithError(errors.New(faker.Sentence()), inPck)
		})
		defer n1.Close()

		n2 := NewJumpNode()
		defer n2.Close()

		n2.Out(node.PortIO).Link(n1.In(node.PortIO))

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

func TestJumpNodeCodec_Decode(t *testing.T) {
	codec := NewJumpNodeCodec()

	spec := &JumpNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func BenchmarkJumpNode_SendAndReceive(b *testing.B) {
	n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})
	defer n1.Close()

	n2 := NewJumpNode()
	defer n2.Close()

	n2.Out(node.PortIO).Link(n1.In(node.PortIO))

	in := port.NewOut()
	in.Link(n2.In(node.PortIn))

	out := port.NewIn()
	n2.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Close()

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		<-inWriter.Receive()
	}
}

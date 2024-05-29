package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewCallNode(t *testing.T) {
	n := NewCallNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestCallNode_Port(t *testing.T) {
	n := NewCallNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestCallNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToMultipleOutputs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n1.Close()

		n2 := NewCallNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		out1 := port.NewIn()
		n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := object.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader1.Read():
			assert.Equal(t, inPayload.Interface(), outPck.Payload().Interface())
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

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.WithError(errors.New(faker.Sentence()))
		})
		defer n1.Close()

		n2 := NewCallNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		err := port.NewIn()
		n2.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := object.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

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

func TestCallNodeCodec_Decode(t *testing.T) {
	codec := NewCallNodeCodec()

	spec := &CallNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkCallNode_SendAndReceive(b *testing.B) {
	n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})
	defer n1.Close()

	n2 := NewCallNode()
	defer n2.Close()

	n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

	in := port.NewOut()
	in.Link(n2.In(node.PortIn))

	out1 := port.NewIn()
	n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader1 := out1.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := object.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		outPck := <-outReader1.Read()
		outReader1.Receive(outPck)

		<-inWriter.Receive()
	}
}

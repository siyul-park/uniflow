package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestNewSequentialNode(t *testing.T) {
	n := NewSequentialNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSequentialNode_Children(t *testing.T) {
	c1 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, nil
	})
	c2 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, nil
	})

	n := NewSequentialNode(c1, c2)
	defer n.Close()

	assert.Equal(t, []node.Node{c1, c2}, n.Children())
}

func TestSequentialNode_Port(t *testing.T) {
	c1 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, nil
	})
	c2 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, nil
	})

	n := NewSequentialNode(c1, c2)
	defer n.Close()

	assert.Equal(t, c1.In(node.PortIO), n.In(node.PortIO))
	assert.Equal(t, c1.In(node.PortIn), n.In(node.PortIn))
	assert.Equal(t, c2.Out(node.PortOut), n.Out(node.PortOut))
	assert.Equal(t, c1.Out(node.PortErr), n.Out(node.PortErr))
}

func TestSequentialNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	c1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})
	c2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})

	n := NewSequentialNode(c1, c2)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Close()

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPayload := primitive.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

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
}

func TestSequentialNodeCodec_Decode(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	codec := NewSequentialNodeCodec(s)

	spec := &SequentialNodeSpec{
		Children: []*scheme.Unstructured{
			scheme.NewUnstructured(primitive.NewMap(
				primitive.NewString(scheme.KeyKind), primitive.NewString(kind),
			)),
		},
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

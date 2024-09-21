package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBlockNodeCodec_Decode(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	c := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	codec := NewBlockNodeCodec(s)

	spec := &BlockNodeSpec{
		Specs: []*spec.Unstructured{
			{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: kind,
				},
			},
		},
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewBlockNode(t *testing.T) {
	n := NewBlockNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestBlockNode_Load(t *testing.T) {
	sb := &symbol.Symbol{
		Node: node.NewOneToOneNode(nil),
	}

	n := NewBlockNode(sb)
	defer n.Close()

	count := 0
	h := symbol.LoadFunc(func(s *symbol.Symbol) error {
		count++
		return nil
	})

	err := n.Load(h)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestBlockNode_Unload(t *testing.T) {
	sb := &symbol.Symbol{
		Node: node.NewOneToOneNode(nil),
	}

	n := NewBlockNode(sb)
	defer n.Close()

	count := 0
	h := symbol.UnloadFunc(func(s *symbol.Symbol) error {
		count++
		return nil
	})

	err := n.Unload(h)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestBlockNode_Port(t *testing.T) {
	n := NewBlockNode(&symbol.Symbol{
		Node: node.NewOneToOneNode(nil),
	})
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestBlockNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToNoOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewBlockNode(
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return inPck, nil
				}),
			},
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return inPck, nil
				}),
			},
		)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewBlockNode(
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return inPck, nil
				}),
			},
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return inPck, nil
				}),
			},
		)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, inPayload, outPck.Payload())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewBlockNode(
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return nil, packet.New(types.NewString(faker.UUIDHyphenated()))
				}),
			},
			&symbol.Symbol{
				Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
					return inPck, nil
				}),
			},
		)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-errReader.Read():
			assert.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkBlockNode_SendAndReceive(b *testing.B) {
	n := NewBlockNode(
		&symbol.Symbol{
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return inPck, nil
			}),
		},
	)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		<-inWriter.Receive()
	}
}

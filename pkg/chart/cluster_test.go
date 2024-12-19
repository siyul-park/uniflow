package chart

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
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewClusterNode(t *testing.T) {
	n := NewClusterNode(nil)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestClusterNode_Keys(t *testing.T) {
	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewClusterNode([]*symbol.Symbol{sb})
	defer n.Close()

	keys := n.Keys()
	assert.Len(t, keys, 1)
	assert.Equal(t, sb.ID(), keys[0])
}

func TestClusterNode_Lookup(t *testing.T) {
	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewClusterNode([]*symbol.Symbol{sb})
	defer n.Close()

	assert.Equal(t, sb, n.Lookup(sb.ID()))
}

func TestClusterNode_Inbound(t *testing.T) {
	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewClusterNode([]*symbol.Symbol{sb})
	defer n.Close()

	n.Inbound(node.PortIn, sb.ID(), node.PortIn)
	assert.NotNil(t, n.In(node.PortIn))
}

func TestClusterNode_Outbound(t *testing.T) {
	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewClusterNode([]*symbol.Symbol{sb})
	defer n.Close()

	n.Outbound(node.PortOut, sb.ID(), node.PortOut)
	assert.NotNil(t, n.Out(node.PortOut))
}

func TestClusterNode_Load(t *testing.T) {
	sb1 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	sb2 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
			Ports: map[string][]spec.Port{
				node.PortOut: {
					{
						ID:   sb1.ID(),
						Port: node.PortIn,
					},
				},
			},
		},
		Node: node.NewOneToOneNode(nil),
	}
	n := NewClusterNode([]*symbol.Symbol{sb1, sb2})
	defer n.Close()

	err := n.Load(nil)
	assert.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	assert.Equal(t, 1, out.Links())
}

func TestClusterNode_Unload(t *testing.T) {
	sb1 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	sb2 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
			Ports: map[string][]spec.Port{
				node.PortOut: {
					{
						ID:   sb1.ID(),
						Port: node.PortIn,
					},
				},
			},
		},
		Node: node.NewOneToOneNode(nil),
	}
	n := NewClusterNode([]*symbol.Symbol{sb1, sb2})
	defer n.Close()

	_ = n.Load(nil)

	err := n.Unload(nil)
	assert.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	assert.Equal(t, 0, out.Links())
}

func TestClusterNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		}),
	}

	n := NewClusterNode([]*symbol.Symbol{sb})
	defer n.Close()

	_ = n.Load(nil)

	n.Inbound(node.PortIn, sb.ID(), node.PortIn)
	n.Outbound(node.PortOut, sb.ID(), node.PortOut)

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
}

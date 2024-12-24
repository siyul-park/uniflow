package symbol

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewCluster(t *testing.T) {
	n := NewCluster(nil)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestCluster_Inbound(t *testing.T) {
	sb := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewCluster([]*Symbol{sb})
	defer n.Close()

	n.Inbound(node.PortIn, spec.Port{
		ID:   sb.ID(),
		Port: node.PortIn,
	})
	assert.NotNil(t, n.In(node.PortIn))
}

func TestCluster_Outbound(t *testing.T) {
	sb := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	n := NewCluster([]*Symbol{sb})
	defer n.Close()

	n.Outbound(node.PortOut, spec.Port{
		ID:   sb.ID(),
		Port: node.PortOut,
	})
	assert.NotNil(t, n.Out(node.PortOut))
}

func TestCluster_Load(t *testing.T) {
	sb1 := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	sb2 := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
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
	n := NewCluster([]*Symbol{sb1, sb2})
	defer n.Close()

	err := n.Load(nil)
	assert.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	assert.Equal(t, 1, out.Links())
}

func TestCluster_Unload(t *testing.T) {
	sb1 := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	sb2 := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
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
	n := NewCluster([]*Symbol{sb1, sb2})
	defer n.Close()

	_ = n.Load(nil)

	err := n.Unload(nil)
	assert.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	assert.Equal(t, 0, out.Links())
}

func TestCluster_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	sb := &Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		}),
	}

	n := NewCluster([]*Symbol{sb})
	defer n.Close()

	_ = n.Load(nil)

	n.Inbound(node.PortIn, spec.Port{
		ID:   sb.ID(),
		Port: node.PortIn,
	})
	n.Outbound(node.PortOut, spec.Port{
		ID:   sb.ID(),
		Port: node.PortOut,
	})

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

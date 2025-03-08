package symbol

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
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNewCluster(t *testing.T) {
	n := NewCluster(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
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
	require.NotNil(t, n.In(node.PortIn))
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
	require.NotNil(t, n.Out(node.PortOut))
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
	require.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	require.Len(t, out.Links(), 1)
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
	require.NoError(t, err)

	out := sb2.Node.Out(node.PortOut)
	require.Len(t, out.Links(), 0)
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
		require.Fail(t, ctx.Err().Error())
	}
}

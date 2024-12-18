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
	n := NewClusterNode(symbol.NewTable())
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestClusterNode_Inbound(t *testing.T) {
	tb := symbol.NewTable()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	tb.Insert(sb)

	n := NewClusterNode(tb)
	defer n.Close()

	n.Inbound(node.PortIn, sb.ID(), node.PortIn)
	assert.NotNil(t, n.In(node.PortIn))
}

func TestClusterNode_Outbound(t *testing.T) {
	tb := symbol.NewTable()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	tb.Insert(sb)

	n := NewClusterNode(tb)
	defer n.Close()

	n.Outbound(node.PortOut, sb.ID(), node.PortOut)
	assert.NotNil(t, n.Out(node.PortOut))
}

func TestClusterNode_Symbols(t *testing.T) {
	tb := symbol.NewTable()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	tb.Insert(sb)

	n := NewClusterNode(tb)
	defer n.Close()

	symbols := n.Symbols()
	assert.Len(t, symbols, 1)
	assert.Equal(t, sb, symbols[0])
}

func TestClusterNode_Symbol(t *testing.T) {
	tb := symbol.NewTable()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	tb.Insert(sb)

	n := NewClusterNode(tb)
	defer n.Close()

	assert.Equal(t, sb, n.Symbol(sb.ID()))
}

func TestClusterNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	tb := symbol.NewTable()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.Word(),
		},
		Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		}),
	}
	tb.Insert(sb)

	n := NewClusterNode(tb)
	defer n.Close()

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

package table

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/require"
)

func TestFrameTable_Scan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	agent := runtime.NewAgent()

	tlb := NewFrameTable(agent)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		}),
	}
	defer sb.Close()

	in := port.NewOut()
	defer in.Close()

	out := port.NewIn()
	defer out.Close()

	in.Link(sb.In(node.PortIn))
	sb.Out(node.PortOut).Link(out)

	agent.Load(sb)
	defer agent.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	pck := packet.New(nil)

	inWriter.Write(pck)
	<-outReader.Read()

	outReader.Receive(pck)
	<-inWriter.Receive()

	cursor, err := tlb.Scan(ctx)
	require.NoError(t, err)

	rows, err := schema.ReadAll(cursor)
	require.NoError(t, err)
	require.Len(t, rows, 2)
}

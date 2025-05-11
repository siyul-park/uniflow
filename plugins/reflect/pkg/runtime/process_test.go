package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/require"
)

func TestProcessTable_Scan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	agent := runtime.NewAgent()

	tlb := NewProcessTable(agent)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	in := sb.In(node.PortIn)
	out := sb.Out(node.PortOut)

	agent.Load(sb)
	defer agent.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	in.Open(proc)
	out.Open(proc)

	cursor, err := tlb.Scan(ctx)
	require.NoError(t, err)

	rows, err := schema.ReadAll(cursor)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

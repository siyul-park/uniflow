package debug

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestNewDebugger(t *testing.T) {
	d := NewDebugger()
	assert.NotNil(t, d)
}

func TestDebugger_Symbol(t *testing.T) {
	d := NewDebugger()

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	d.Load(sym)
	defer d.Unload(sym)

	_, ok := d.Symbol(sym.ID())
	assert.True(t, ok)

	ids := d.Symbols()
	assert.Contains(t, ids, sym.ID())
}

func TestDebugger_Process(t *testing.T) {
	d := NewDebugger()

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}

	in := sym.In(node.PortIn)

	d.Load(sym)
	defer d.Unload(sym)

	proc := process.New()
	defer proc.Exit(nil)

	in.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		default:
			if _, ok := d.Process(proc.ID()); ok {
				ids := d.Processes()
				assert.Contains(t, ids, proc.ID())
				return
			}
		}
	}
}

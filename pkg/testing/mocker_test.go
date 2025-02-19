package testing

import (
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMocker_Mock(t *testing.T) {
	m := NewMocker()

	sb1 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb1.Close()

	sb2 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb2.Close()

	sb3 := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb3.Close()

	sb1.Spec.SetPorts(map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   sb2.ID(),
				Port: node.PortIn,
			},
		},
	})
	sb2.Spec.SetPorts(map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   sb3.ID(),
				Port: node.PortIn,
			},
		},
	})

	sb1.Out(node.PortOut).Link(sb2.In(node.PortIn))
	sb2.Out(node.PortOut).Link(sb3.In(node.PortIn))

	m.Load(sb1)
	defer m.Unload(sb1)

	m.Load(sb2)
	defer m.Unload(sb2)

	m.Load(sb3)
	defer m.Unload(sb3)

	proc := process.New()
	defer proc.Exit(nil)

	sb4 := &symbol.Symbol{
		Spec: sb2.Spec,
		Node: node.NewOneToOneNode(nil),
	}
	defer sb4.Close()

	err := m.Mock(proc, sb4)
	assert.NoError(t, err)

	out := sb1.Out(node.PortOut)
	writer := out.Open(proc)
	assert.Len(t, writer.Links(), 1)

	out = sb4.Out(node.PortOut)
	writer = out.Open(proc)
	assert.Len(t, writer.Links(), 1)
}

package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestSymbol_Getter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Annotations: map[string]string{
			faker.UUIDHyphenated(): faker.UUIDHyphenated(),
		},
		Ports: map[string][]spec.Port{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
	}

	sym := &Symbol{
		Spec: meta,
		Node: n,
	}

	assert.Equal(t, meta.GetID(), sym.ID())
	assert.Equal(t, meta.GetKind(), sym.Kind())
	assert.Equal(t, meta.GetNamespace(), sym.Namespace())
	assert.Equal(t, meta.GetName(), sym.Name())
	assert.Equal(t, meta.GetAnnotations(), sym.Annotations())

	p1 := n.In(node.PortIn)
	p2 := sym.In(node.PortIn)

	assert.Equal(t, p1, p2)
}

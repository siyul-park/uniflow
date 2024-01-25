package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestSymbol_Getter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Links: map[string][]scheme.PortLocation{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
	}

	sym := New(spec, n)

	assert.Equal(t, spec.GetID(), sym.ID())
	assert.Equal(t, spec.GetKind(), sym.Kind())
	assert.Equal(t, spec.GetNamespace(), sym.Namespace())
	assert.Equal(t, spec.GetName(), sym.Name())

	p1, _ := n.Port(node.PortIn)
	p2, _ := sym.Port(node.PortIn)

	assert.Equal(t, p1, p2)
}

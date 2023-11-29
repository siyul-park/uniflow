package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestSymbol_Getter(t *testing.T) {
	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n.Close()
	spec := &scheme.SpecMeta{
		ID:        n.ID(),
		Kind:      faker.Word(),
		Namespace: scheme.NamespaceDefault,
		Name:      faker.UUIDHyphenated(),
		Links: map[string][]scheme.PortLocation{
			node.PortOut: {
				{
					ID:   ulid.Make(),
					Port: node.PortIn,
				},
			},
		},
	}

	sym := &Symbol{Node: n, Spec: spec}

	assert.Equal(t, n.ID(), sym.ID())
	assert.Equal(t, spec.GetKind(), sym.Kind())
	assert.Equal(t, spec.GetNamespace(), sym.Namespace())
	assert.Equal(t, spec.GetName(), sym.Name())
	assert.Equal(t, spec.GetLinks(), sym.Links())

	p1, _ := n.Port(node.PortIn)
	p2, _ := sym.Port(node.PortIn)

	assert.Equal(t, p1, p2)
}
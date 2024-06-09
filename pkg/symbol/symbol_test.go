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

	spec := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Annotations: map[string]string{
			faker.UUIDHyphenated(): faker.UUIDHyphenated(),
		},
		Links: map[string][]spec.PortLocation{
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
	assert.Equal(t, spec.GetAnnotations(), sym.Annotations())
	assert.Equal(t, StatusNotReady, sym.Status())
	assert.Equal(t, spec, sym.Spec())
	assert.Equal(t, n, sym.Unwrap())

	p1 := n.In(node.PortIn)
	p2 := sym.In(node.PortIn)

	assert.Equal(t, p1, p2)
}

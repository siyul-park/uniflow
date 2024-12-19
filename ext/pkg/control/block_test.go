package control

import (
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockNodeCodec_Compile(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	c := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	codec := NewBlockNodeCodec(s)

	sp := &BlockNodeSpec{
		Specs: []spec.Spec{
			&spec.Unstructured{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: kind,
				},
			},
			&spec.Unstructured{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: kind,
				},
			},
		},
		Inbound: map[string][]spec.Port{
			node.PortIn: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
		Outbound: map[string][]spec.Port{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortOut,
				},
			},
		},
	}

	n, err := codec.Compile(sp)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

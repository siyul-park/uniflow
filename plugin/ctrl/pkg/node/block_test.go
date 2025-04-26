package node

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/require"
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
		Specs: []*spec.Unstructured{
			{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: kind,
				},
			},
			{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: kind,
				},
			},
		},
		Inbounds: map[string][]spec.Port{
			node.PortIn: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
		Outbounds: map[string][]spec.Port{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortOut,
				},
			},
		},
	}

	n, err := codec.Compile(sp)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

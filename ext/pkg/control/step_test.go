package control

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestStepNodeCodec_Compile(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	c := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	codec := NewStepNodeCodec(s)

	sp := &StepNodeSpec{
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
	}

	n, err := codec.Compile(sp)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

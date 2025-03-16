package spec

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSpecDecoder_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[types.Value, any]()
	dec.Add(newSpecDecoder(types.Decoder))

	unstructured := &Unstructured{
		Meta: Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Fields: map[string]any{},
	}
	v, _ := types.Marshal(unstructured)

	var decoded Spec
	err := dec.Decode(v, &decoded)
	require.NoError(t, err)

	require.Equal(t, unstructured, decoded)
}

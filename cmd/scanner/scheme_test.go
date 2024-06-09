package scanner

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestSpecCodec_Decode(t *testing.T) {
	s := spec.NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	codec := NewSpecCodec(SpecCodecOptions{
		Scheme: s,
	})

	data := map[string]any{
		spec.KeyID:   uuid.Must(uuid.NewV7()).String(),
		spec.KeyKind: kind,
	}

	meta, err := codec.Decode(data)
	assert.NoError(t, err)
	assert.IsType(t, meta, &spec.Meta{})
	assert.Equal(t, data[spec.KeyID], meta.GetID().String())
	assert.Equal(t, data[spec.KeyKind], meta.GetKind())
}

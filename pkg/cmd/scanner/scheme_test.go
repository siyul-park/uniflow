package scanner

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestSpecCodec_Decode(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})

	codec := NewSpecCodec(SpecCodecOptions{
		Scheme: s,
	})

	data := map[string]any{
		scheme.KeyID:   uuid.Must(uuid.NewV7()).String(),
		scheme.KeyKind: kind,
	}

	spec, err := codec.Decode(data)
	assert.NoError(t, err)
	assert.IsType(t, spec, &scheme.SpecMeta{})
	assert.Equal(t, data[scheme.KeyID], spec.GetID().String())
	assert.Equal(t, data[scheme.KeyKind], spec.GetKind())
}

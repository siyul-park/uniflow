package scheme

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestUnstructured_GetAndSetID(t *testing.T) {
	id := ulid.Make()

	u := NewUnstructured(nil)

	u.SetID(id)
	assert.Equal(t, id, u.GetID())
}

func TestUnstructured_GetAndSetKind(t *testing.T) {
	kind := faker.UUIDHyphenated()

	u := NewUnstructured(nil)

	u.SetKind(kind)
	assert.Equal(t, kind, u.GetKind())
}

func TestUnstructured_GetAndNamespace(t *testing.T) {
	namespace := faker.UUIDHyphenated()

	u := NewUnstructured(nil)

	u.SetNamespace(namespace)
	assert.Equal(t, namespace, u.GetNamespace())
}

func TestUnstructured_GetAndLinks(t *testing.T) {
	links := map[string][]PortLocation{
		faker.UUIDHyphenated(): {
			{
				ID:   ulid.Make(),
				Port: faker.UUIDHyphenated(),
			},
		},
	}

	u := NewUnstructured(nil)

	u.SetLinks(links)
	assert.Equal(t, links, u.GetLinks())
}

func TestUnstructured_Marshal(t *testing.T) {
	u := NewUnstructured(nil)
	spec := &SpecMeta{
		ID:   ulid.Make(),
		Kind: faker.UUIDHyphenated(),
	}

	err := u.Marshal(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.GetID(), u.GetID())
	assert.Equal(t, spec.GetKind(), u.GetKind())
}

func TestUnstructured_Unmarshal(t *testing.T) {
	u := NewUnstructured(nil)
	spec := &SpecMeta{}

	_ = u.Marshal(&SpecMeta{
		ID:   ulid.Make(),
		Kind: faker.UUIDHyphenated(),
	})

	err := u.Unmarshal(spec)
	assert.NoError(t, err)
	assert.Equal(t, u.GetID(), spec.GetID())
	assert.Equal(t, u.GetKind(), spec.GetKind())
}

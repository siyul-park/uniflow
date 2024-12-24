package resource

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMeta_GetSet(t *testing.T) {
	meta := &Meta{
		ID:          uuid.Must(uuid.NewV7()),
		Namespace:   "default",
		Name:        faker.UUIDHyphenated(),
		Annotations: map[string]string{"key": "value"},
	}

	assert.Equal(t, meta.ID, meta.GetID())
	assert.Equal(t, meta.Namespace, meta.GetNamespace())
	assert.Equal(t, meta.Name, meta.GetName())
	assert.Equal(t, meta.Annotations, meta.GetAnnotations())
}

func TestMatch(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())

	sp := &Meta{ID: id1, Namespace: "default", Name: "node1"}
	examples := []*Meta{
		{ID: id1, Namespace: "default", Name: "node1"},
		{ID: id1},
		{Namespace: "default", Name: "node1"},
		{ID: id2, Namespace: "default", Name: "node2"},
		{ID: id2},
		{Namespace: "default", Name: "node2"},
	}

	expeced := []*Meta{examples[0], examples[1], examples[2]}

	assert.Equal(t, expeced, Match(sp, examples...))
}

package spec

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMeta_GetSet(t *testing.T) {
	meta := &Meta{
		ID:          uuid.Must(uuid.NewV7()),
		Kind:        faker.Word(),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Ports:       map[string][]Port{"out": {{Name: faker.Word(), Port: "in"}}},
		Env:         map[string][]Secret{"env1": {{Name: "secret1", Value: "value1"}}},
	}

	assert.Equal(t, meta.ID, meta.GetID())
	assert.Equal(t, meta.Kind, meta.GetKind())
	assert.Equal(t, meta.Namespace, meta.GetNamespace())
	assert.Equal(t, meta.Name, meta.GetName())
	assert.Equal(t, meta.Annotations, meta.GetAnnotations())
	assert.Equal(t, meta.Ports, meta.GetPorts())
	assert.Equal(t, meta.Env, meta.GetEnv())
}

func TestMatch(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())

	spc := &Meta{ID: id1, Namespace: "default", Name: "node1"}
	examples := []Spec{
		&Meta{ID: id1, Namespace: "default", Name: "node1"},
		&Meta{ID: id1},
		&Meta{Namespace: "default", Name: "node1"},
		&Meta{ID: id2, Namespace: "default", Name: "node2"},
		&Meta{ID: id2},
		&Meta{Namespace: "default", Name: "node2"},
	}

	expeced := []Spec{examples[0], examples[1], examples[2]}

	assert.Equal(t, expeced, Match(spc, examples...))
}

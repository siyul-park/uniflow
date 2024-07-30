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

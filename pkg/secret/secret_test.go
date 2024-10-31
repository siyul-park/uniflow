package secret

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSecret_Get(t *testing.T) {
	scrt := &Secret{
		ID:          uuid.Must(uuid.NewV7()),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Data:        faker.Word(),
	}

	assert.Equal(t, scrt.ID, scrt.GetID())
	assert.Equal(t, scrt.Namespace, scrt.GetNamespace())
	assert.Equal(t, scrt.Name, scrt.GetName())
	assert.Equal(t, scrt.Annotations, scrt.GetAnnotations())
	assert.Equal(t, scrt.Data, scrt.GetData())
	assert.True(t, scrt.IsIdentified())
}

func TestSecret_Set(t *testing.T) {
	scrt := New()

	id := uuid.Must(uuid.NewV7())
	namespace := "default"
	name := faker.Word()
	annotations := map[string]string{"key": "value"}
	data := faker.Word()

	scrt.SetID(id)
	scrt.SetNamespace(namespace)
	scrt.SetName(name)
	scrt.SetAnnotations(annotations)
	scrt.SetData(data)

	assert.Equal(t, id, scrt.GetID())
	assert.Equal(t, namespace, scrt.GetNamespace())
	assert.Equal(t, name, scrt.GetName())
	assert.Equal(t, annotations, scrt.GetAnnotations())
	assert.Equal(t, data, scrt.GetData())
	assert.True(t, scrt.IsIdentified())
}

package secret

import (
	"github.com/go-faker/faker/v4"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSecret_SetID(t *testing.T) {
	scrt := New()
	id := uuid.Must(uuid.NewV7())
	scrt.SetID(id)
	assert.Equal(t, id, scrt.GetID())
}

func TestSecret_SetNamespace(t *testing.T) {
	scrt := New()
	namespace := faker.Word()
	scrt.SetNamespace(namespace)
	assert.Equal(t, namespace, scrt.GetNamespace())
}

func TestSecret_SetName(t *testing.T) {
	scrt := New()
	name := faker.Word()
	scrt.SetName(name)
	assert.Equal(t, name, scrt.GetName())
}

func TestSecret_SetAnnotations(t *testing.T) {
	scrt := New()
	annotation := map[string]string{"key": "value"}
	scrt.SetAnnotations(annotation)
	assert.Equal(t, annotation, scrt.GetAnnotations())
}

func TestSecret_SetData_Nil(t *testing.T) {
	scrt := New()
	data := faker.Word()
	scrt.SetData(data)
	assert.Equal(t, data, scrt.GetData())
}

func TestSecret_IsIdentified(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		scrt := &Secret{
			ID: uuid.Must(uuid.NewV7()),
		}
		assert.True(t, scrt.IsIdentified())
	})

	t.Run("Name", func(t *testing.T) {
		scrt := &Secret{
			Name: faker.Word(),
		}
		assert.True(t, scrt.IsIdentified())
	})

	t.Run("Nil", func(t *testing.T) {
		scrt := &Secret{}
		assert.False(t, scrt.IsIdentified())
	})
}

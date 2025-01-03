package value

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValue_ID(t *testing.T) {
	scrt := New()
	id := uuid.Must(uuid.NewV7())
	scrt.SetID(id)
	assert.Equal(t, id, scrt.GetID())
}

func TestValue_Namespace(t *testing.T) {
	scrt := New()
	namespace := faker.UUIDHyphenated()
	scrt.SetNamespace(namespace)
	assert.Equal(t, namespace, scrt.GetNamespace())
}

func TestValue_Name(t *testing.T) {
	scrt := New()
	name := faker.UUIDHyphenated()
	scrt.SetName(name)
	assert.Equal(t, name, scrt.GetName())
}

func TestValue_Annotations(t *testing.T) {
	scrt := New()
	annotation := map[string]string{"key": "value"}
	scrt.SetAnnotations(annotation)
	assert.Equal(t, annotation, scrt.GetAnnotations())
}

func TestValue_Data(t *testing.T) {
	scrt := New()
	data := faker.UUIDHyphenated()
	scrt.SetData(data)
	assert.Equal(t, data, scrt.GetData())
}

func TestValue_IsIdentified(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		scrt := &Value{
			ID: uuid.Must(uuid.NewV7()),
		}
		assert.True(t, scrt.IsIdentified())
	})

	t.Run("Name", func(t *testing.T) {
		scrt := &Value{
			Name: faker.UUIDHyphenated(),
		}
		assert.True(t, scrt.IsIdentified())
	})

	t.Run("Nil", func(t *testing.T) {
		scrt := &Value{}
		assert.False(t, scrt.IsIdentified())
	})
}

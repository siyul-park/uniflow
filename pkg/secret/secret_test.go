package secret

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSecret_GetSet(t *testing.T) {
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
}

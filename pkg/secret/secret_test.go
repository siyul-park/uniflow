package secret

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSecret_GetSet(t *testing.T) {
	sec := &Secret{
		ID:          uuid.Must(uuid.NewV7()),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Data:        faker.Word(),
	}

	assert.Equal(t, sec.ID, sec.GetID())
	assert.Equal(t, sec.Namespace, sec.GetNamespace())
	assert.Equal(t, sec.Name, sec.GetName())
	assert.Equal(t, sec.Annotations, sec.GetAnnotations())
	assert.Equal(t, sec.Data, sec.GetData())
}

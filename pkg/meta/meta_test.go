package meta

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestNamespacedName(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		unstructured := &Unstructured{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: "default",
		}
		require.Equal(t, unstructured.GetNamespace()+"/"+unstructured.GetID().String(), NamespacedName(unstructured))
	})
	t.Run("Name", func(t *testing.T) {
		unstructured := &Unstructured{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: "default",
			Name:      faker.UUIDHyphenated(),
		}
		require.Equal(t, unstructured.GetNamespace()+"/"+unstructured.GetName(), NamespacedName(unstructured))
	})
}

package meta

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnstructured_GetAndSet(t *testing.T) {
	t.Run("KeyID", func(t *testing.T) {
		unstructured := &Unstructured{}
		id := uuid.Must(uuid.NewV7())
		unstructured.Set(KeyID, id)
		val, ok := unstructured.Get(KeyID)
		require.True(t, ok)
		require.Equal(t, id, val)
	})

	t.Run("KeyNamespace", func(t *testing.T) {
		unstructured := &Unstructured{}
		unstructured.Set(KeyNamespace, "default")
		val, ok := unstructured.Get(KeyNamespace)
		require.True(t, ok)
		require.Equal(t, "default", val)
	})

	t.Run("KeyName", func(t *testing.T) {
		unstructured := &Unstructured{}
		name := faker.UUIDHyphenated()
		unstructured.Set(KeyName, name)
		val, ok := unstructured.Get(KeyName)
		require.True(t, ok)
		require.Equal(t, name, val)
	})

	t.Run("KeyAnnotations", func(t *testing.T) {
		unstructured := &Unstructured{}
		annotations := map[string]string{"key": "value"}
		unstructured.Set(KeyAnnotations, annotations)
		val, ok := unstructured.Get(KeyAnnotations)
		require.True(t, ok)
		require.Equal(t, annotations, val)
	})

	t.Run("CustomField", func(t *testing.T) {
		unstructured := &Unstructured{}
		customKey := "customField"
		customValue := "customValue"
		unstructured.Set(customKey, customValue)
		val, ok := unstructured.Get(customKey)
		require.True(t, ok)
		require.Equal(t, customValue, val)
	})
}

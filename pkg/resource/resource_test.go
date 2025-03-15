package resource

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestMeta_ID(t *testing.T) {
	meta := &Meta{}
	id := uuid.Must(uuid.NewV7())
	meta.SetID(id)
	require.Equal(t, id, meta.GetID())
}

func TestMeta_Namespace(t *testing.T) {
	meta := &Meta{}
	namespace := "default"
	meta.SetNamespace(namespace)
	require.Equal(t, namespace, meta.GetNamespace())
}

func TestMeta_Name(t *testing.T) {
	meta := &Meta{}
	name := faker.UUIDHyphenated()
	meta.SetName(name)
	require.Equal(t, name, meta.GetName())
}

func TestMeta_Annotations(t *testing.T) {
	meta := &Meta{}
	annotations := map[string]string{"key": "value"}
	meta.SetAnnotations(annotations)
	require.Equal(t, annotations, meta.GetAnnotations())
}

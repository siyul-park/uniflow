package resource

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	id := uuid.Must(uuid.NewV7())

	tests := []struct {
		source Resource
		target Resource
		expect bool
	}{
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{ID: id},
			expect: true,
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{ID: uuid.Must(uuid.NewV7())},
			expect: false,
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{Namespace: DefaultNamespace},
			expect: true,
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{Namespace: "other"},
			expect: false,
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{Name: "node1"},
			expect: true,
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			target: &Meta{Name: "node2"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v, %v", tt.source, tt.target), func(t *testing.T) {
			assert.Equal(t, tt.expect, Is(tt.source, tt.target))
		})
	}
}

func TestMeta_ID(t *testing.T) {
	meta := &Meta{}
	id := uuid.Must(uuid.NewV7())
	meta.SetID(id)
	assert.Equal(t, id, meta.GetID())
}

func TestMeta_Namespace(t *testing.T) {
	meta := &Meta{}
	namespace := "default"
	meta.SetNamespace(namespace)
	assert.Equal(t, namespace, meta.GetNamespace())
}

func TestMeta_Name(t *testing.T) {
	meta := &Meta{}
	name := faker.UUIDHyphenated()
	meta.SetName(name)
	assert.Equal(t, name, meta.GetName())
}

func TestMeta_Annotations(t *testing.T) {
	meta := &Meta{}
	annotations := map[string]string{"key": "value"}
	meta.SetAnnotations(annotations)
	assert.Equal(t, annotations, meta.GetAnnotations())
}

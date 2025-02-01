package resource

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	id := uuid.Must(uuid.NewV7())

	tests := []struct {
		source   *Meta
		examples []*Meta
		matches  []int
	}{
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			examples: []*Meta{
				{
					ID: id,
				},
				{
					ID: uuid.Must(uuid.NewV7()),
				},
			},
			matches: []int{0},
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			examples: []*Meta{
				{
					Namespace: DefaultNamespace,
				},
				{
					Namespace: "other",
				},
			},
			matches: []int{0},
		},
		{
			source: &Meta{
				ID:        id,
				Namespace: DefaultNamespace,
				Name:      "node1",
			},
			examples: []*Meta{
				{
					Name: "node1",
				},
				{
					Name: "node2",
				},
			},
			matches: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v, %v", tt.source, tt.examples), func(t *testing.T) {
			expected := make([]*Meta, 0, len(tt.matches))
			for _, i := range tt.matches {
				expected = append(expected, tt.examples[i])
			}
			assert.Equal(t, expected, Match(tt.source, tt.examples...))
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

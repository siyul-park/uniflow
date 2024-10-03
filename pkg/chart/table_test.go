package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(chrt.GetID()))
}

func TestTable_Free(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt)
	assert.NoError(t, err)

	ok, err := tb.Free(chrt.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestTable_Lookup(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt)
	assert.NoError(t, err)
	assert.Equal(t, chrt, tb.Lookup(chrt.GetID()))
}

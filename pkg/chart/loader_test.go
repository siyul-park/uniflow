package chart

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/assert"
)

func TestLoader_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chartStore := NewStore()
	valueStore := value.NewStore()

	table := NewTable()
	defer table.Close()

	loader := NewLoader(LoaderConfig{
		Table:      table,
		ChartStore: chartStore,
		ValueStore: valueStore,
	})

	scrt := &value.Value{
		ID:   uuid.Must(uuid.NewV7()),
		Data: faker.UUIDHyphenated(),
	}

	chrt1 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),

		Env: map[string][]spec.Value{
			"key": {
				{
					ID:   scrt.GetID(),
					Data: faker.UUIDHyphenated(),
				},
			},
		},
	}
	chrt2 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []*spec.Unstructured{
			{
				Meta: spec.Meta{
					Kind:      chrt1.GetName(),
					Namespace: resource.DefaultNamespace,
					Name:      faker.UUIDHyphenated(),
				},
			},
		},
	}

	valueStore.Store(ctx, scrt)

	chartStore.Store(ctx, chrt1)
	chartStore.Store(ctx, chrt2)

	err := loader.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.NotNil(t, table.Lookup(chrt1.GetID()))
	assert.NotNil(t, table.Lookup(chrt2.GetID()))
}

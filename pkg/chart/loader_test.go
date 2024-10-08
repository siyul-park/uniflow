package chart

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestLoader_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chartStore := NewStore()
	secretStore := secret.NewStore()

	table := NewTable()
	defer table.Close()

	loader := NewLoader(LoaderConfig{
		Table:       table,
		ChartStore:  chartStore,
		SecretStore: secretStore,
	})

	sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
	chrt1 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs:     []spec.Spec{},
		Env: map[string][]Value{
			"key": {
				{
					ID:    sec.GetID(),
					Value: faker.Word(),
				},
			},
		},
	}
	chrt2 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind:      chrt1.GetName(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	secretStore.Store(ctx, sec)

	chartStore.Store(ctx, chrt1)
	chartStore.Store(ctx, chrt2)

	err := loader.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.NotNil(t, table.Lookup(chrt1.GetID()))
	assert.NotNil(t, table.Lookup(chrt2.GetID()))
}

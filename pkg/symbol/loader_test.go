package symbol

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestLoader_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("LoadMultipleSpecs", func(t *testing.T) {
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta1 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string][]spec.Secret{
				"key": {
					{
						ID:    sec.GetID(),
						Value: faker.Word(),
					},
				},
			},
		}
		meta2 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports:     map[string][]spec.Port{node.PortIO: {{ID: meta1.GetID(), Port: node.PortIO}}},
		}
		meta3 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports:     map[string][]spec.Port{node.PortIO: {{Name: meta2.GetName(), Port: node.PortIO}}},
		}

		secretStore.Store(ctx, sec)
		specStore.Store(ctx, meta1)
		specStore.Store(ctx, meta2)
		specStore.Store(ctx, meta3)

		err := loader.Load(ctx, meta1, meta2, meta3)
		assert.NoError(t, err)

		_, ok := table.Lookup(meta1.GetID())
		assert.True(t, ok)

		_, ok = table.Lookup(meta2.GetID())
		assert.True(t, ok)

		_, ok = table.Lookup(meta3.GetID())
		assert.True(t, ok)
	})

	t.Run("ReloadWithSameID", func(t *testing.T) {
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string][]spec.Secret{
				"key": {
					{
						ID:    sec.GetID(),
						Value: faker.Word(),
					},
				},
			},
		}

		secretStore.Store(ctx, sec)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		assert.NoError(t, err)

		err = loader.Load(ctx, meta)
		assert.NoError(t, err)
	})

	t.Run("ReloadAfterDelete", func(t *testing.T) {
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string][]spec.Secret{
				"key": {
					{
						ID:    sec.GetID(),
						Value: faker.Word(),
					},
				},
			},
		}

		secretStore.Store(ctx, sec)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		assert.NoError(t, err)

		specStore.Delete(ctx, meta)

		err = loader.Load(ctx, meta)
		assert.NoError(t, err)

		_, ok := table.Lookup(meta.GetID())
		assert.False(t, ok)
	})

	t.Run("LoadMultipleSecrets", func(t *testing.T) {
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		sec1 := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		sec2 := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string][]spec.Secret{
				"sec1": {
					{
						ID:    sec1.GetID(),
						Value: faker.Word(),
					},
				},
				"sec2": {
					{
						ID:    sec2.GetID(),
						Value: faker.Word(),
					},
				},
			},
		}

		secretStore.Store(ctx, sec1)
		secretStore.Store(ctx, sec2)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		assert.NoError(t, err)

		_, ok := table.Lookup(meta.GetID())
		assert.True(t, ok)
	})

	t.Run("LoadNonExistSecret", func(t *testing.T) {
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string][]spec.Secret{
				"nonexist": {
					{
						ID:    uuid.Must(uuid.NewV7()),
						Value: faker.Word(),
					},
				},
			},
		}

		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		assert.Error(t, err)

		_, ok := table.Lookup(meta.GetID())
		assert.False(t, ok)
	})
}

package symbol

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/require"
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
		valueStore := value.NewStore()

		table := NewTable()
		defer table.Close()

		loader := NewLoader(LoaderConfig{
			Table:      table,
			Scheme:     s,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}

		meta1 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string]spec.Value{
				"key": {
					ID:   scrt.GetID(),
					Data: faker.UUIDHyphenated(),
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

		valueStore.Store(ctx, scrt)
		specStore.Store(ctx, meta1)
		specStore.Store(ctx, meta2)
		specStore.Store(ctx, meta3)

		err := loader.Load(ctx, meta1, meta2, meta3)
		require.NoError(t, err)
		require.NotNil(t, table.Lookup(meta1.GetID()))
		require.NotNil(t, table.Lookup(meta2.GetID()))
		require.NotNil(t, table.Lookup(meta3.GetID()))
	})

	t.Run("ReloadWithSameID", func(t *testing.T) {
		specStore := spec.NewStore()
		valueStore := value.NewStore()

		table := NewTable()
		defer table.Close()

		loader := NewLoader(LoaderConfig{
			Table:      table,
			Scheme:     s,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string]spec.Value{
				"key": {
					ID:   scrt.GetID(),
					Data: faker.UUIDHyphenated(),
				},
			},
		}

		valueStore.Store(ctx, scrt)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		require.NoError(t, err)

		err = loader.Load(ctx, meta)
		require.NoError(t, err)
	})

	t.Run("ReloadAfterDelete", func(t *testing.T) {
		specStore := spec.NewStore()
		valueStore := value.NewStore()

		table := NewTable()
		defer table.Close()

		loader := NewLoader(LoaderConfig{
			Table:      table,
			Scheme:     s,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string]spec.Value{
				"key": {
					ID:   scrt.GetID(),
					Data: faker.UUIDHyphenated(),
				},
			},
		}

		valueStore.Store(ctx, scrt)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		require.NoError(t, err)

		specStore.Delete(ctx, meta)

		err = loader.Load(ctx, meta)
		require.NoError(t, err)
		require.Nil(t, table.Lookup(meta.GetID()))
	})

	t.Run("LoadMultipleValues", func(t *testing.T) {
		specStore := spec.NewStore()
		valueStore := value.NewStore()

		table := NewTable()
		defer table.Close()

		loader := NewLoader(LoaderConfig{
			Table:      table,
			Scheme:     s,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		sec1 := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}
		sec2 := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string]spec.Value{
				"sec1": {
					ID:   sec1.GetID(),
					Data: faker.UUIDHyphenated(),
				},
				"sec2": {
					ID:   sec2.GetID(),
					Data: faker.UUIDHyphenated(),
				},
			},
		}

		valueStore.Store(ctx, sec1)
		valueStore.Store(ctx, sec2)
		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		require.NoError(t, err)
		require.NotNil(t, table.Lookup(meta.GetID()))
	})

	t.Run("LoadNonExistValue", func(t *testing.T) {
		specStore := spec.NewStore()
		valueStore := value.NewStore()

		table := NewTable()
		defer table.Close()

		loader := NewLoader(LoaderConfig{
			Table:      table,
			Scheme:     s,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string]spec.Value{
				"nonexist": {
					ID:   uuid.Must(uuid.NewV7()),
					Data: faker.UUIDHyphenated(),
				},
			},
		}

		specStore.Store(ctx, meta)

		err := loader.Load(ctx, meta)
		require.Error(t, err)
	})
}

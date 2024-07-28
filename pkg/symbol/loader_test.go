package symbol

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
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
		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta1 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env:       map[string]spec.Secret{"": {ID: sec.GetID()}},
		}
		meta2 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports:     map[string][]spec.Port{node.PortIO: {{ID: meta1.GetID(), Port: node.PortIO}}},
		}
		meta3 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports:     map[string][]spec.Port{node.PortIO: {{Name: meta2.GetName(), Port: node.PortIO}}},
		}

		scStore.Store(ctx, sec)
		spStore.Store(ctx, meta1)
		spStore.Store(ctx, meta2)
		spStore.Store(ctx, meta3)

		res, err := loader.Load(ctx, meta1, meta2, meta3)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		_, ok := table.Lookup(meta1.GetID())
		assert.True(t, ok)

		_, ok = table.Lookup(meta2.GetID())
		assert.True(t, ok)

		_, ok = table.Lookup(meta3.GetID())
		assert.True(t, ok)
	})

	t.Run("ReloadWithSameID", func(t *testing.T) {
		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Env:       map[string]spec.Secret{"": {ID: sec.GetID()}},
		}

		scStore.Store(ctx, sec)
		spStore.Store(ctx, meta)

		res1, err := loader.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, res1)

		res2, err := loader.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, res2)

		assert.Equal(t, res1, res2)
	})

	t.Run("ReloadAfterDelete", func(t *testing.T) {
		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Env:       map[string]spec.Secret{"": {ID: sec.GetID()}},
		}

		scStore.Store(ctx, sec)
		spStore.Store(ctx, meta)

		res1, err := loader.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, res1)

		spStore.Delete(ctx, meta)

		res2, err := loader.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Nil(t, res2)

		_, ok := table.Lookup(meta.GetID())
		assert.False(t, ok)
	})

	t.Run("LoadMultipleSecrets", func(t *testing.T) {
		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		sec1 := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		sec2 := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string]spec.Secret{
				"sec1": {ID: sec1.GetID()},
				"sec2": {ID: sec2.GetID()},
			},
		}

		scStore.Store(ctx, sec1)
		scStore.Store(ctx, sec2)
		spStore.Store(ctx, meta)

		res, err := loader.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		_, ok := table.Lookup(meta.GetID())
		assert.True(t, ok)
	})

	t.Run("LoadNonExistSecret", func(t *testing.T) {
		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Env: map[string]spec.Secret{
				"nonexist": {ID: uuid.Must(uuid.NewV7())},
			},
		}

		spStore.Store(ctx, meta)

		_, err := loader.Load(ctx, meta)
		assert.NoError(t, err)

		_, ok := table.Lookup(meta.GetID())
		assert.True(t, ok)
	})
}

func TestLoader_Reconcile(t *testing.T) {
	t.Run("ReconcileLoadedSpec", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		err := loader.Watch(ctx)
		assert.NoError(t, err)

		go loader.Reconcile(ctx)

		sec := &secret.Secret{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.Word(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Env: map[string]spec.Secret{
				"": {
					ID:    sec.GetID(),
					Value: "{{ . }}",
				},
			},
		}

		scStore.Store(ctx, sec)
		spStore.Store(ctx, meta)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sym, ok := table.Lookup(meta.GetID()); ok {
						assert.Equal(t, meta.GetID(), sym.ID())
						assert.Equal(t, sec.Data, sym.Env()[""].Value)
						return
					}

				}
			}
		}()

		spStore.Delete(ctx, meta)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if _, ok := table.Lookup(meta.GetID()); !ok {
						return
					}
				}
			}
		}()
	})

	t.Run("ReconcileLoadedSecret", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		spStore := spec.NewStore()
		scStore := secret.NewStore()

		table := NewTable()
		defer table.Clear()

		loader := NewLoader(LoaderConfig{
			Table:       table,
			Scheme:      s,
			SpecStore:   spStore,
			SecretStore: scStore,
		})
		defer loader.Close()

		err := loader.Watch(ctx)
		assert.NoError(t, err)

		go loader.Reconcile(ctx)

		sec := &secret.Secret{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.Word(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Env: map[string]spec.Secret{
				"": {
					ID:    sec.GetID(),
					Value: "{{ . }}",
				},
			},
		}

		spStore.Store(ctx, meta)
		scStore.Store(ctx, sec)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sym, ok := table.Lookup(meta.GetID()); ok {
						assert.Equal(t, meta.GetID(), sym.ID())
						assert.Equal(t, sec.Data, sym.Env()[""].Value)
						return
					}

				}
			}
		}()

		scStore.Delete(ctx, sec)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sym, ok := table.Lookup(meta.GetID()); ok {
						if sec.Data != sym.Env()[""].Value {
							return
						}
					}
				}
			}
		}()
	})
}

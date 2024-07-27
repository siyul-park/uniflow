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
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Load", func(t *testing.T) {
		spst := spec.NewStore()
		scst := secret.NewStore()

		tb := NewTable()
		defer tb.Clear()

		ld := NewLoader(LoaderConfig{
			Table:       tb,
			Scheme:      s,
			SpecStore:   spst,
			SecretStore: scst,
		})
		defer ld.Close()

		meta1 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}
		meta2 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports: map[string][]spec.Port{
				node.PortIO: {
					{
						ID:   meta1.GetID(),
						Port: node.PortIO,
					},
				},
			},
		}
		meta3 := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Ports: map[string][]spec.Port{
				node.PortIO: {
					{
						Name: meta2.GetName(),
						Port: node.PortIO,
					},
				},
			},
		}

		spst.Store(ctx, meta1)
		spst.Store(ctx, meta2)
		spst.Store(ctx, meta3)

		r, err := ld.Load(ctx, meta3)
		assert.NoError(t, err)
		assert.NotNil(t, r)

		_, ok := tb.Lookup(meta1.GetID())
		assert.True(t, ok)

		_, ok = tb.Lookup(meta2.GetID())
		assert.True(t, ok)

		_, ok = tb.Lookup(meta3.GetID())
		assert.True(t, ok)

	})

	t.Run("Reload Same ID", func(t *testing.T) {
		spst := spec.NewStore()
		scst := secret.NewStore()

		tb := NewTable()
		defer tb.Clear()

		ld := NewLoader(LoaderConfig{
			Table:       tb,
			Scheme:      s,
			SpecStore:   spst,
			SecretStore: scst,
		})
		defer ld.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		spst.Store(ctx, meta)

		r1, err := ld.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		r2, err := ld.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		assert.Equal(t, r1, r2)
	})

	t.Run("Reload After Delete", func(t *testing.T) {
		spst := spec.NewStore()
		scst := secret.NewStore()

		tb := NewTable()
		defer tb.Clear()

		ld := NewLoader(LoaderConfig{
			Table:       tb,
			Scheme:      s,
			SpecStore:   spst,
			SecretStore: scst,
		})
		defer ld.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		spst.Store(ctx, meta)

		r1, err := ld.Load(ctx, meta)
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		spst.Delete(ctx, meta)

		r2, err := ld.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Nil(t, r2)

		_, ok := tb.Lookup(meta.GetID())
		assert.False(t, ok)
	})
}

func TestLoader_Reconcile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	spst := spec.NewStore()
	scst := secret.NewStore()

	tb := NewTable()
	defer tb.Clear()

	ld := NewLoader(LoaderConfig{
		Table:       tb,
		Scheme:      s,
		SpecStore:   spst,
		SecretStore: scst,
	})
	defer ld.Close()

	err := ld.Watch(ctx)
	assert.NoError(t, err)

	go ld.Reconcile(ctx)

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	spst.Store(ctx, meta)

	func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			default:
				if sym, ok := tb.Lookup(meta.GetID()); ok {
					assert.Equal(t, meta.GetID(), sym.ID())
					return
				}
			}
		}
	}()
}

func BenchmarkLoader_Load(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	spst := spec.NewStore()
	scst := secret.NewStore()

	tb := NewTable()
	defer tb.Clear()

	ld := NewLoader(LoaderConfig{
		Table:       tb,
		Scheme:      s,
		SpecStore:   spst,
		SecretStore: scst,
	})
	defer ld.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	spst.Store(ctx, meta)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, err := ld.Load(ctx, meta)
		assert.NoError(b, err)
		assert.NotNil(b, r)

		b.StopTimer()

		tb.Free(meta.GetID())

		b.StartTimer()
	}
}

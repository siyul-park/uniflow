package loader

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestLoader_LoadOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Load", func(t *testing.T) {
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		spec1 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}
		spec2 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						ID:   spec1.GetID(),
						Port: node.PortIO,
					},
				},
			},
		}
		spec3 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						Name: spec2.GetName(),
						Port: node.PortIO,
					},
				},
			},
		}

		st.InsertOne(ctx, spec1)
		st.InsertOne(ctx, spec2)
		st.InsertOne(ctx, spec3)

		r, err := ld.LoadOne(ctx, spec3.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r)

		_, ok := tb.LookupByID(spec1.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(spec2.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(spec3.GetID())
		assert.True(t, ok)

	})

	t.Run("Reload Same ID", func(t *testing.T) {
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		spec := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		st.InsertOne(ctx, spec)

		r1, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		r2, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		assert.Equal(t, r1, r2)
	})

	t.Run("Reload After Delete", func(t *testing.T) {
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		spec := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		st.InsertOne(ctx, spec)

		r1, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		st.DeleteOne(ctx, storage.Where[uuid.UUID](scheme.KeyID).EQ(spec.GetID()))

		r2, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.Nil(t, r2)

		_, ok := tb.LookupByID(spec.GetID())
		assert.False(t, ok)
	})
}

func TestLoader_LoadAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Load", func(t *testing.T) {
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		spec1 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}
		spec2 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						ID:   spec1.GetID(),
						Port: node.PortIO,
					},
				},
			},
		}
		spec3 := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						Name: spec2.GetName(),
						Port: node.PortIO,
					},
				},
			},
		}

		st.InsertOne(ctx, spec1)
		st.InsertOne(ctx, spec2)
		st.InsertOne(ctx, spec3)

		r, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, r, 3)

		sym1, ok := tb.LookupByID(spec1.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym1)

		sym2, ok := tb.LookupByID(spec2.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym2)

		sym3, ok := tb.LookupByID(spec3.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym3)
	})

	t.Run("Reload", func(t *testing.T) {
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		spec := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		st.InsertOne(ctx, spec)

		r1, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, r1, 1)

		r2, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.NotEqual(t, r1, r2)
	})
}

func BenchmarkLoader_LoadOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Storage: st,
		Table:   tb,
	})

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	st.InsertOne(ctx, spec)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(b, err)
		assert.NotNil(b, r)

		b.StopTimer()

		tb.Free(spec.GetID())

		b.StartTimer()
	}
}

func BenchmarkLoader_LoadAll(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Storage: st,
		Table:   tb,
	})

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	st.InsertOne(ctx, spec)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, err := ld.LoadAll(ctx)
		assert.NoError(b, err)
		assert.Len(b, r, 1)

		b.StopTimer()

		tb.Free(spec.GetID())

		b.StartTimer()
	}
}

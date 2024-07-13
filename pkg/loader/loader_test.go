package loader

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestLoader_LoadOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Load", func(t *testing.T) {
		st := store.New(memdb.NewCollection(""))

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Store: st,
			Table: tb,
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
			Links: map[string][]spec.PortLocation{
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
			Links: map[string][]spec.PortLocation{
				node.PortIO: {
					{
						Name: meta2.GetName(),
						Port: node.PortIO,
					},
				},
			},
		}

		st.InsertOne(ctx, meta1)
		st.InsertOne(ctx, meta2)
		st.InsertOne(ctx, meta3)

		r, err := ld.LoadOne(ctx, meta3.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r)

		_, ok := tb.LookupByID(meta1.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(meta2.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(meta3.GetID())
		assert.True(t, ok)

	})

	t.Run("Reload Same ID", func(t *testing.T) {
		st := store.New(memdb.NewCollection(""))

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Store: st,
			Table: tb,
		})
		defer ld.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		st.InsertOne(ctx, meta)

		r1, err := ld.LoadOne(ctx, meta.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		r2, err := ld.LoadOne(ctx, meta.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		assert.Equal(t, r1, r2)
	})

	t.Run("Reload After Delete", func(t *testing.T) {
		st := store.New(memdb.NewCollection(""))

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Store: st,
			Table: tb,
		})
		defer ld.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		st.InsertOne(ctx, meta)

		r1, err := ld.LoadOne(ctx, meta.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		st.DeleteOne(ctx, store.Where[uuid.UUID](spec.KeyID).Equal(meta.GetID()))

		r2, err := ld.LoadOne(ctx, meta.GetID())
		assert.NoError(t, err)
		assert.Nil(t, r2)

		_, ok := tb.LookupByID(meta.GetID())
		assert.False(t, ok)
	})
}

func TestLoader_LoadAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Load", func(t *testing.T) {
		st := store.New(memdb.NewCollection(""))

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Store: st,
			Table: tb,
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
			Links: map[string][]spec.PortLocation{
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
			Links: map[string][]spec.PortLocation{
				node.PortIO: {
					{
						Name: meta2.GetName(),
						Port: node.PortIO,
					},
				},
			},
		}

		st.InsertOne(ctx, meta1)
		st.InsertOne(ctx, meta2)
		st.InsertOne(ctx, meta3)

		r, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, r, 3)

		sym1, ok := tb.LookupByID(meta1.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym1)

		sym2, ok := tb.LookupByID(meta2.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym2)

		sym3, ok := tb.LookupByID(meta3.GetID())
		assert.True(t, ok)
		assert.Contains(t, r, sym3)
	})

	t.Run("Reload", func(t *testing.T) {
		st := store.New(memdb.NewCollection(""))

		tb := symbol.NewTable(s)
		defer tb.Clear()

		ld := New(Config{
			Store: st,
			Table: tb,
		})
		defer ld.Close()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		st.InsertOne(ctx, meta)

		r1, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, r1, 1)

		r2, err := ld.LoadAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, r2, 1)

		assert.False(t, r1[0] == r2[0])
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

	st := store.New(memdb.NewCollection(""))

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Namespace: spec.DefaultNamespace,
		Store:     st,
		Table:     tb,
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

	st.InsertOne(ctx, meta)

	func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			default:
				if sym, ok := tb.LookupByID(meta.GetID()); ok {
					assert.Equal(t, meta.GetID(), sym.ID())
					return
				}
			}
		}
	}()
}

func BenchmarkLoader_LoadOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := store.New(memdb.NewCollection(""))

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Store: st,
		Table: tb,
	})

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	st.InsertOne(ctx, meta)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, err := ld.LoadOne(ctx, meta.GetID())
		assert.NoError(b, err)
		assert.NotNil(b, r)

		b.StopTimer()

		tb.Free(meta.GetID())

		b.StartTimer()
	}
}

func BenchmarkLoader_LoadAll(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := store.New(memdb.NewCollection(""))

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Store: st,
		Table: tb,
	})

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	st.InsertOne(ctx, meta)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, err := ld.LoadAll(ctx)
		assert.NoError(b, err)
		assert.Len(b, r, 1)

		b.StopTimer()

		tb.Free(meta.GetID())

		b.StartTimer()
	}
}

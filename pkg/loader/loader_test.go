package loader

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
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

	t.Run("load first", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Clear() }()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec1 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}
		spec2 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						ID:   spec1.GetID(),
						Port: node.PortIO,
					},
				},
			},
		}
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		st.InsertOne(ctx, spec1)
		st.InsertOne(ctx, spec2)

		r, err := ld.LoadOne(ctx, spec2.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r)

		_, ok := tb.LookupByID(spec1.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(spec2.GetID())
		assert.True(t, ok)

	})

	t.Run("reload same ID", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Clear() }()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(ctx, spec)

		r1, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		r2, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		assert.Equal(t, r1, r2)
	})

	t.Run("reload after deletion", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(ctx, storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Clear() }()

		ld := New(Config{
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(ctx, spec)

		r1, err := ld.LoadOne(ctx, spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		st.DeleteOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))

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
	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable(s)
	defer func() { _ = tb.Clear() }()

	ld := New(Config{
		Storage: st,
		Table:   tb,
	})

	kind := faker.Word()

	spec1 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}
	spec2 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Links: map[string][]scheme.PortLocation{
			node.PortIO: {
				{
					ID:   spec1.GetID(),
					Port: node.PortIO,
				},
			},
		},
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	st.InsertOne(ctx, spec1)
	st.InsertOne(ctx, spec2)

	r, err := ld.LoadAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, r, 2)

	sym1, ok := tb.LookupByID(spec1.GetID())
	assert.True(t, ok)
	assert.Contains(t, r, sym1)

	sym2, ok := tb.LookupByID(spec2.GetID())
	assert.True(t, ok)
	assert.Contains(t, r, sym2)

	r, err = ld.LoadAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, r, 2)
	assert.NotContains(t, r, sym1)
	assert.NotContains(t, r, sym2)
}

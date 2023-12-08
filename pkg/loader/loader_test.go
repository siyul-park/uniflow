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
	t.Run("load first", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Close() }()

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

		st.InsertOne(context.Background(), spec1)
		st.InsertOne(context.Background(), spec2)

		r, err := ld.LoadOne(context.Background(), spec2.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r)

		_, ok := tb.LookupByID(spec1.GetID())
		assert.True(t, ok)

		_, ok = tb.LookupByID(spec2.GetID())
		assert.True(t, ok)

	})

	t.Run("reload same ID", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Close() }()

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

		st.InsertOne(context.Background(), spec)

		r1, err := ld.LoadOne(context.Background(), spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		r2, err := ld.LoadOne(context.Background(), spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		assert.Equal(t, r1, r2)
	})

	t.Run("reload after deletion", func(t *testing.T) {
		s := scheme.New()
		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable(s)
		defer func() { _ = tb.Close() }()

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

		st.InsertOne(context.Background(), spec)

		r1, err := ld.LoadOne(context.Background(), spec.GetID())
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		st.DeleteOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))

		r2, err := ld.LoadOne(context.Background(), spec.GetID())
		assert.NoError(t, err)
		assert.Nil(t, r2)

		_, ok := tb.LookupByID(spec.GetID())
		assert.False(t, ok)
	})
}

func TestLoader_LoadAll(t *testing.T) {
	s := scheme.New()
	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable(s)
	defer func() { _ = tb.Close() }()

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

	st.InsertOne(context.Background(), spec1)
	st.InsertOne(context.Background(), spec2)

	r, err := ld.LoadAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, r, 2)

	_, ok := tb.LookupByID(spec1.GetID())
	assert.True(t, ok)

	_, ok = tb.LookupByID(spec2.GetID())
	assert.True(t, ok)
}

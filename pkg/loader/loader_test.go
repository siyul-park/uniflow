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
	t.Run("linked all", func(t *testing.T) {
		s := scheme.New()

		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable()
		defer func() { _ = tb.Close() }()

		ld, _ := New(context.Background(), Config{
			Scheme:  s,
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec1 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
		}
		spec2 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
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
			return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(context.Background(), spec1)
		st.InsertOne(context.Background(), spec2)

		r2, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec2.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		n1, ok := tb.Lookup(spec1.GetID())
		assert.True(t, ok)

		n2, ok := tb.Lookup(spec2.GetID())
		assert.True(t, ok)

		p1, _ := n1.Port(node.PortIO)
		p2, _ := n2.Port(node.PortIO)

		assert.Equal(t, p1.Links(), 1)
		assert.Equal(t, p2.Links(), 1)
	})

	t.Run("linked all with name", func(t *testing.T) {
		s := scheme.New()

		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable()
		defer func() { _ = tb.Close() }()

		ld, _ := New(context.Background(), Config{
			Scheme:  s,
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec1 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
			Name:      faker.Word(),
		}
		spec2 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
			Name:      faker.Word(),
			Links: map[string][]scheme.PortLocation{
				node.PortIO: {
					{
						Name: spec1.Name,
						Port: node.PortIO,
					},
				},
			},
		}

		codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(context.Background(), spec1)
		st.InsertOne(context.Background(), spec2)

		r2, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec2.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		n1, ok := tb.Lookup(spec1.GetID())
		assert.True(t, ok)

		n2, ok := tb.Lookup(spec2.GetID())
		assert.True(t, ok)

		p1, _ := n1.Port(node.PortIO)
		p2, _ := n2.Port(node.PortIO)

		assert.Equal(t, p1.Links(), 1)
		assert.Equal(t, p2.Links(), 1)
	})

	t.Run("unlinked any", func(t *testing.T) {
		s := scheme.New()

		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable()
		defer func() { _ = tb.Close() }()

		ld, _ := New(context.Background(), Config{
			Scheme:  s,
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec1 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
		}
		spec2 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
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
			return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(context.Background(), spec2)

		r2, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec2.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		st.InsertOne(context.Background(), spec1)

		r1, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec1.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		n1, ok := tb.Lookup(spec1.GetID())
		assert.True(t, ok)

		n2, ok := tb.Lookup(spec2.GetID())
		assert.True(t, ok)

		p1, _ := n1.Port(node.PortIO)
		p2, _ := n2.Port(node.PortIO)

		assert.Equal(t, p1.Links(), 1)
		assert.Equal(t, p2.Links(), 1)
	})

	t.Run("relink any", func(t *testing.T) {
		s := scheme.New()

		st, _ := storage.New(context.Background(), storage.Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		tb := symbol.NewTable()
		defer func() { _ = tb.Close() }()

		ld, _ := New(context.Background(), Config{
			Scheme:  s,
			Storage: st,
			Table:   tb,
		})

		kind := faker.Word()

		spec1 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
		}
		spec2 := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.NamespaceDefault,
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
			return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
		})

		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, codec)

		st.InsertOne(context.Background(), spec1)
		st.InsertOne(context.Background(), spec2)

		r2, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec2.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r2)

		ok, err := ld.UnloadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec1.GetID()))
		assert.NoError(t, err)
		assert.True(t, ok)

		r1, err := ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec1.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, r1)

		n1, ok := tb.Lookup(spec1.GetID())
		assert.True(t, ok)

		n2, ok := tb.Lookup(spec2.GetID())
		assert.True(t, ok)

		p1, _ := n1.Port(node.PortIO)
		p2, _ := n2.Port(node.PortIO)

		assert.Equal(t, p1.Links(), 1)
		assert.GreaterOrEqual(t, p2.Links(), 1)
	})
}

func TestLoader_LoadMany(t *testing.T) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld, _ := New(context.Background(), Config{
		Scheme:  s,
		Storage: st,
		Table:   tb,
	})

	kind := faker.Word()

	spec1 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.NamespaceDefault,
	}
	spec2 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.NamespaceDefault,
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
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	st.InsertOne(context.Background(), spec1)
	st.InsertOne(context.Background(), spec2)

	r, err := ld.LoadMany(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, r, 2)

	_, ok := tb.Lookup(spec1.GetID())
	assert.True(t, ok)

	_, ok = tb.Lookup(spec2.GetID())
	assert.True(t, ok)
}

func TestLoader_UnloadOne(t *testing.T) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld, _ := New(context.Background(), Config{
		Scheme:  s,
		Storage: st,
		Table:   tb,
	})

	kind := faker.Word()

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.NamespaceDefault,
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	st.InsertOne(context.Background(), spec)

	_, _ = ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))

	ok, err := ld.UnloadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.True(t, ok)

	_, ok = tb.Lookup(spec.GetID())
	assert.False(t, ok)
}

func TestLoader_UnloadMany(t *testing.T) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld, _ := New(context.Background(), Config{
		Scheme:  s,
		Storage: st,
		Table:   tb,
	})

	kind := faker.Word()

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.NamespaceDefault,
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	st.InsertOne(context.Background(), spec)

	_, _ = ld.LoadOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))

	count, err := ld.UnloadMany(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	_, ok := tb.Lookup(spec.GetID())
	assert.False(t, ok)
}

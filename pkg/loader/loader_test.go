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
	s := scheme.New()
	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld := New(Config{
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

	r, err := ld.LoadOne(context.Background(), spec2.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, r)

	_, ok := tb.LookupByID(spec1.GetID())
	assert.True(t, ok)

	_, ok = tb.LookupByID(spec2.GetID())
	assert.True(t, ok)

	st.DeleteOne(context.Background(), storage.Where[ulid.ULID](scheme.KeyID).EQ(spec2.GetID()))

	r, err = ld.LoadOne(context.Background(), spec2.GetID())
	assert.NoError(t, err)
	assert.Nil(t, r)

	_, ok = tb.LookupByID(spec2.GetID())
	assert.False(t, ok)
}

func TestLoader_LoadAll(t *testing.T) {
	s := scheme.New()
	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld := New(Config{
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

	r, err := ld.LoadAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, r, 2)

	_, ok := tb.LookupByID(spec1.GetID())
	assert.True(t, ok)

	_, ok = tb.LookupByID(spec2.GetID())
	assert.True(t, ok)
}

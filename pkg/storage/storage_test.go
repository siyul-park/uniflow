package storage

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestStorage_Watch(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	stream, err := st.Watch(context.Background(), nil)
	assert.NoError(t, err)
	defer func() { _ = stream.Close() }()

	go func() {
		for {
			event, ok := <-stream.Next()
			if ok {
				assert.NotNil(t, event.NodeID)
			} else {
				return
			}
		}
	}()

	_, _ = st.InsertOne(context.Background(), spec)
	_, _ = st.UpdateOne(context.Background(), spec)
	_, _ = st.DeleteOne(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
}

func TestStorage_InsertOne(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	id, err := st.InsertOne(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.ID, id)
}

func TestStorage_InsertMany(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := []scheme.Spec{
		&scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		},
		&scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		},
	}

	ids, err := st.InsertMany(context.Background(), spec)
	assert.NoError(t, err)
	assert.Len(t, ids, len(spec))
	for i, spec := range spec {
		assert.Equal(t, spec.GetID(), ids[i])
	}
}

func TestStorage_UpdateOne(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	ok, err := st.UpdateOne(context.Background(), spec)
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(context.Background(), spec)

	ok, err = st.UpdateOne(context.Background(), spec)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestStorage_UpdateMany(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := []scheme.Spec{
		&scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		},
		&scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		},
	}

	count, err := st.UpdateMany(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertMany(context.Background(), spec)

	count, err = st.UpdateMany(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, len(spec), count)
}

func TestStorage_DeleteOne(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	ok, err := st.DeleteOne(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(context.Background(), spec)

	ok, err = st.DeleteOne(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestStorage_DeleteMany(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	count, err := st.DeleteMany(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertOne(context.Background(), spec)

	count, err = st.DeleteMany(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestStorage_FindOne(t *testing.T) {
	t.Run("id", func(t *testing.T) {
		kind := faker.Word()

		s := scheme.New()
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{
				ID: spec.GetID(),
			}), nil
		}))

		st, _ := New(context.Background(), Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		spec := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		_, _ = st.InsertOne(context.Background(), spec)

		def, err := st.FindOne(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
		assert.NoError(t, err)
		assert.NotNil(t, def)
		assert.Equal(t, spec.GetID(), def.GetID())
	})

	t.Run("namespace, name", func(t *testing.T) {
		kind := faker.Word()

		s := scheme.New()
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{
				ID: spec.GetID(),
			}), nil
		}))

		st, _ := New(context.Background(), Config{
			Scheme:   s,
			Database: memdb.New(faker.Word()),
		})

		spec := &scheme.SpecMeta{
			ID:        ulid.Make(),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.Word(),
		}

		_, _ = st.InsertOne(context.Background(), spec)

		def, err := st.FindOne(context.Background(), Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace()).And(Where[string](scheme.KeyName).EQ(spec.GetName())))
		assert.NoError(t, err)
		assert.NotNil(t, def)
		assert.Equal(t, spec.GetID(), def.GetID())
	})
}

func TestStorage_FindMany(t *testing.T) {
	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{
			ID: spec.GetID(),
		}), nil
	}))

	st, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	_, _ = st.InsertOne(context.Background(), spec)

	defs, err := st.FindMany(context.Background(), Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.Len(t, defs, 1)
	assert.Equal(t, spec.GetID(), defs[0].GetID())
}

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

const batchSize = 100

func TestNewStorage(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		kind := faker.UUIDHyphenated()

		s := scheme.New()
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		st, err := New(ctx, Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		})
		assert.NoError(t, err)
		assert.NotNil(t, st)
	})

	t.Run("Load", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		kind := faker.UUIDHyphenated()

		s := scheme.New()
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		config := Config{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		}

		_, _ = New(ctx, config)

		st, err := New(ctx, config)
		assert.NoError(t, err)
		assert.NotNil(t, st)
	})
}

func TestStorage_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	stream, err := st.Watch(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	defer stream.Close()

	go func() {
		for {
			if event, ok := <-stream.Next(); ok {
				assert.NotNil(t, event.NodeID)
			} else {
				return
			}
		}
	}()

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, spec)
	_, _ = st.UpdateOne(ctx, spec)
	_, _ = st.DeleteOne(ctx, Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
}

func TestStorage_InsertOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	id, err := st.InsertOne(ctx, spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.GetID(), id)
}

func TestStorage_InsertMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var specs []scheme.Spec
	for i := 0; i < batchSize; i++ {
		specs = append(specs, &scheme.SpecMeta{
			ID:   ulid.Make(),
			Kind: kind,
		})
	}

	ids, err := st.InsertMany(ctx, specs)
	assert.NoError(t, err)
	assert.Len(t, ids, len(specs))
	for i, spec := range specs {
		assert.Equal(t, spec.GetID(), ids[i])
	}
}

func TestStorage_UpdateOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	id := ulid.Make()

	origin := &scheme.SpecMeta{
		ID:        id,
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	patch := &scheme.SpecMeta{
		ID:        id,
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	ok, err := st.UpdateOne(ctx, patch)
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(ctx, origin)

	ok, err = st.UpdateOne(ctx, patch)
	assert.NoError(t, err)
	assert.True(t, ok)

	res, err := st.FindOne(ctx, Where[ulid.ULID](scheme.KeyID).EQ(id))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, patch)
}

func TestStorage_UpdateMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []ulid.ULID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, ulid.Make())
	}

	var origins []scheme.Spec
	var patches []scheme.Spec
	for _, id := range ids {
		origins = append(origins, &scheme.SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
		patches = append(patches, &scheme.SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	count, err := st.UpdateMany(ctx, patches)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertMany(ctx, origins)

	count, err = st.UpdateMany(ctx, patches)
	assert.NoError(t, err)
	assert.Equal(t, len(patches), count)

	res, err := st.FindMany(ctx, Where[ulid.ULID](scheme.KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, len(patches))
	for _, patch := range patches {
		assert.Contains(t, res, patch)
	}
}

func TestStorage_DeleteOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	ok, err := st.DeleteOne(ctx, Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(ctx, spec)

	ok, err = st.DeleteOne(ctx, Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestStorage_DeleteMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []ulid.ULID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, ulid.Make())
	}

	var specs []scheme.Spec
	for _, id := range ids {
		specs = append(specs, &scheme.SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	count, err := st.DeleteMany(ctx, Where[ulid.ULID](scheme.KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertMany(ctx, specs)

	count, err = st.DeleteMany(ctx, Where[ulid.ULID](scheme.KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, len(specs), count)
}

func TestStorage_FindOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, spec)

	def, err := st.FindOne(ctx, Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, spec.GetID(), def.GetID())
}

func TestStorage_FindMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := New(ctx, Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []ulid.ULID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, ulid.Make())
	}

	var specs []scheme.Spec
	for _, id := range ids {
		specs = append(specs, &scheme.SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, specs)

	res, err := st.FindMany(ctx, Where[ulid.ULID](scheme.KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, len(specs))
	for _, spec := range specs {
		assert.Contains(t, res, spec)
	}
}

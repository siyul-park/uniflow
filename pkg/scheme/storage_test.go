package scheme

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

const batchSize = 100

func TestNewStorage(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		kind := faker.UUIDHyphenated()

		s := New()
		s.AddKnownType(kind, &SpecMeta{})
		s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		st, err := NewStorage(ctx, StorageConfig{
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

		s := New()
		s.AddKnownType(kind, &SpecMeta{})
		s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		config := StorageConfig{
			Scheme:   s,
			Database: memdb.New(faker.UUIDHyphenated()),
		}

		_, _ = NewStorage(ctx, config)

		st, err := NewStorage(ctx, config)
		assert.NoError(t, err)
		assert.NotNil(t, st)
	})
}

func TestStorage_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
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

	spec := &SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, spec)
	_, _ = st.UpdateOne(ctx, spec)
	_, _ = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
}

func TestStorage_InsertOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
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

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var specs []Spec
	for i := 0; i < batchSize; i++ {
		specs = append(specs, &SpecMeta{
			ID:   uuid.Must(uuid.NewV7()),
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

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	id := uuid.Must(uuid.NewV7())

	origin := &SpecMeta{
		ID:        id,
		Kind:      kind,
		Namespace: DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	patch := &SpecMeta{
		ID:        id,
		Kind:      kind,
		Namespace: DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	ok, err := st.UpdateOne(ctx, patch)
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(ctx, origin)

	ok, err = st.UpdateOne(ctx, patch)
	assert.NoError(t, err)
	assert.True(t, ok)

	res, err := st.FindOne(ctx, Where[uuid.UUID](KeyID).EQ(id))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, patch)
}

func TestStorage_UpdateMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var origins []Spec
	var patches []Spec
	for _, id := range ids {
		origins = append(origins, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
		patches = append(patches, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
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

	res, err := st.FindMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
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

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	ok, err := st.DeleteOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(ctx, spec)

	ok, err = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestStorage_DeleteMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var specs []Spec
	for _, id := range ids {
		specs = append(specs, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	count, err := st.DeleteMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertMany(ctx, specs)

	count, err = st.DeleteMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, len(specs), count)
}

func TestStorage_FindOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, spec)

	def, err := st.FindOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, spec.GetID(), def.GetID())
}

func TestStorage_FindMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var specs []Spec
	for _, id := range ids {
		specs = append(specs, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, specs)

	res, err := st.FindMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, len(specs))
	for _, spec := range specs {
		assert.Contains(t, res, spec)
	}
}

func BenchmarkStorage_InsertOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		spec := &SpecMeta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
		}

		_, _ = st.InsertOne(ctx, spec)
	}
}

func BenchmarkStorage_InsertMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var specs []Spec
		for j := 0; j < batchSize; j++ {
			specs = append(specs, &SpecMeta{
				ID:   uuid.Must(uuid.NewV7()),
				Kind: kind,
			})
		}

		_, _ = st.InsertMany(ctx, specs)
	}
}

func BenchmarkStorage_UpdateOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	id := uuid.Must(uuid.NewV7())

	origin := &SpecMeta{
		ID:        id,
		Kind:      kind,
		Namespace: DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	_, _ = st.InsertOne(ctx, origin)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		patch := &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		_, _ = st.UpdateOne(ctx, patch)
	}
}

func BenchmarkStorage_UpdateMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var origins []Spec
	for _, id := range ids {
		origins = append(origins, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, origins)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		var patches []Spec
		for _, id := range ids {
			patches = append(patches, &SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			})
		}

		b.StartTimer()

		_, _ = st.UpdateMany(ctx, patches)
	}
}

func BenchmarkStorage_DeleteOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		spec := &SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: DefaultNamespace,
		}
		_, _ = st.InsertOne(ctx, spec)

		b.StartTimer()

		_, _ = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
	}
}

func BenchmarkStorage_DeleteMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		var ids []uuid.UUID
		for i := 0; i < batchSize; i++ {
			ids = append(ids, uuid.Must(uuid.NewV7()))
		}

		var specs []Spec
		for _, id := range ids {
			specs = append(specs, &SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			})
		}

		_, _ = st.InsertMany(ctx, specs)

		b.StartTimer()

		_, _ = st.DeleteMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
	}

}

func BenchmarkStorage_FindOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	spec := &SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	for i := 0; i < batchSize; i++ {
		_, _ = st.InsertOne(ctx, &SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: DefaultNamespace,
		})
	}
	_, _ = st.InsertOne(ctx, spec)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = st.FindOne(ctx, Where[uuid.UUID](KeyID).EQ(spec.GetID()))
	}
}

func BenchmarkStorage_FindMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := New()
	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := NewStorage(ctx, StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var specs []Spec
	for _, id := range ids {
		specs = append(specs, &SpecMeta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, specs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = st.FindMany(ctx, Where[uuid.UUID](KeyID).IN(ids...))
	}
}

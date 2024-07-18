package spec

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/stretchr/testify/assert"
)

const batchSize = 100

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore(memdb.NewCollection(""))

	err := st.Index(ctx)
	assert.NoError(t, err)

	err = st.Index(ctx)
	assert.NoError(t, err)
}

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

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

	meta := &Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, meta)
	_, _ = st.UpdateOne(ctx, meta)
	_, _ = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
}

func TestStore_InsertOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	id, err := st.InsertOne(ctx, meta)
	assert.NoError(t, err)
	assert.Equal(t, meta.GetID(), id)
}

func TestStore_InsertMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var specs []Spec
	for i := 0; i < batchSize; i++ {
		specs = append(specs, &Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
		})
	}

	ids, err := st.InsertMany(ctx, specs)
	assert.NoError(t, err)
	assert.Len(t, ids, len(specs))
	for i, meta := range specs {
		assert.Equal(t, meta.GetID(), ids[i])
	}
}

func TestStore_UpdateOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	id := uuid.Must(uuid.NewV7())

	origin := &Meta{
		ID:        id,
		Kind:      kind,
		Namespace: DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	patch := &Meta{
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
}

func TestStore_UpdateMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var origins []Spec
	var patches []Spec
	for _, id := range ids {
		origins = append(origins, &Meta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
		patches = append(patches, &Meta{
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
}

func TestStore_DeleteOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta := &Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	ok, err := st.DeleteOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = st.InsertOne(ctx, meta)

	ok, err = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestStore_DeleteMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var spcs []Spec
	for _, id := range ids {
		spcs = append(spcs, &Meta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	count, err := st.DeleteMany(ctx, Where[uuid.UUID](KeyID).In(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = st.InsertMany(ctx, spcs)

	count, err = st.DeleteMany(ctx, Where[uuid.UUID](KeyID).In(ids...))
	assert.NoError(t, err)
	assert.Equal(t, len(spcs), count)
}

func TestStore_FindOne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta := &Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	_, _ = st.InsertOne(ctx, meta)

	def, err := st.FindOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, meta.GetID(), def.GetID())
}

func TestStore_FindMany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var specs []Spec
	for _, id := range ids {
		specs = append(specs, &Meta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, specs)

	res, err := st.FindMany(ctx, Where[uuid.UUID](KeyID).In(ids...))
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, len(specs))
}

func BenchmarkStore_InsertOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		spec := &Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
		}

		_, _ = st.InsertOne(ctx, spec)
	}
}

func BenchmarkStore_InsertMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var specs []Spec
		for j := 0; j < batchSize; j++ {
			specs = append(specs, &Meta{
				ID:   uuid.Must(uuid.NewV7()),
				Kind: kind,
			})
		}

		_, _ = st.InsertMany(ctx, specs)
	}
}

func BenchmarkStore_UpdateOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	id := uuid.Must(uuid.NewV7())

	origin := &Meta{
		ID:        id,
		Kind:      kind,
		Namespace: DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	_, _ = st.InsertOne(ctx, origin)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		patch := &Meta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		_, _ = st.UpdateOne(ctx, patch)
	}
}

func BenchmarkStore_UpdateMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var origins []Spec
	for _, id := range ids {
		origins = append(origins, &Meta{
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
			patches = append(patches, &Meta{
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

func BenchmarkStore_DeleteOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		meta := &Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: DefaultNamespace,
		}
		_, _ = st.InsertOne(ctx, meta)

		b.StartTimer()

		_, _ = st.DeleteOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
	}
}

func BenchmarkStore_DeleteMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		var ids []uuid.UUID
		for i := 0; i < batchSize; i++ {
			ids = append(ids, uuid.Must(uuid.NewV7()))
		}

		var specs []Spec
		for _, id := range ids {
			specs = append(specs, &Meta{
				ID:        id,
				Kind:      kind,
				Namespace: DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			})
		}

		_, _ = st.InsertMany(ctx, specs)

		b.StartTimer()

		_, _ = st.DeleteMany(ctx, Where[uuid.UUID](KeyID).In(ids...))
	}

}

func BenchmarkStore_FindOne(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta := &Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	for i := 0; i < batchSize; i++ {
		_, _ = st.InsertOne(ctx, &Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: DefaultNamespace,
		})
	}
	_, _ = st.InsertOne(ctx, meta)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = st.FindOne(ctx, Where[uuid.UUID](KeyID).Equal(meta.GetID()))
	}
}

func BenchmarkStore_FindMany(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	var ids []uuid.UUID
	for i := 0; i < batchSize; i++ {
		ids = append(ids, uuid.Must(uuid.NewV7()))
	}

	var specs []Spec
	for _, id := range ids {
		specs = append(specs, &Meta{
			ID:        id,
			Kind:      kind,
			Namespace: DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		})
	}

	_, _ = st.InsertMany(ctx, specs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = st.FindMany(ctx, Where[uuid.UUID](KeyID).In(ids...))
	}
}

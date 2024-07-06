package databasetest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

const batchSize = 100
const benchSize = 1000

func TestCollection_Name(t *testing.T, collection database.Collection) {
	t.Helper()

	name := collection.Name()
	assert.NotEmpty(t, name)
}

func TestCollection_Indexes(t *testing.T, collection database.Collection) {
	t.Helper()

	indexes := collection.Indexes()
	assert.NotNil(t, indexes)
}

func TestCollection_Watch(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	stream, err := collection.Watch(ctx, nil)
	assert.NoError(t, err)
	defer stream.Close()

	go func() {
		for {
			event, ok := <-stream.Next()
			if !ok {
				return
			}

			assert.NotNil(t, event.DocumentID)
		}
	}()

	doc := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("version"), types.NewInt64(0),
	)

	_, _ = collection.InsertOne(ctx, doc)
	_, _ = collection.UpdateOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)), types.NewMap(types.NewString("version"), types.NewInt64(1)))
	_, _ = collection.DeleteOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
}

func TestCollection_InsertOne(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()

		doc := types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("version"), types.NewInt64(0),
			types.NewString("deleted"), types.False,
		)

		id, err := collection.InsertOne(ctx, doc)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), id)
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()

		doc := types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("version"), types.NewInt64(0),
			types.NewString("deleted"), types.False,
		)

		_, _ = collection.InsertOne(ctx, doc)

		_, err := collection.InsertOne(ctx, doc)
		assert.Error(t, err)
	})
}

func TestCollection_InsertMany(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()

		var docs []types.Map
		for i := 0; i < batchSize; i++ {
			docs = append(docs, types.NewMap(
				types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
				types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
				types.NewString("version"), types.NewInt64(0),
				types.NewString("deleted"), types.False,
			))
		}

		ids, err := collection.InsertMany(ctx, docs)
		assert.NoError(t, err)
		assert.Len(t, ids, len(docs))
		for i, doc := range docs {
			assert.Equal(t, ids[i], doc.GetOr(types.NewString("id"), nil))
		}
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()

		var docs []types.Map
		for i := 0; i < batchSize; i++ {
			docs = append(docs, types.NewMap(
				types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
				types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
				types.NewString("version"), types.NewInt64(0),
				types.NewString("deleted"), types.False,
			))
		}

		_, _ = collection.InsertMany(ctx, docs)

		_, err := collection.InsertMany(ctx, docs)
		assert.Error(t, err)
	})
}

func TestCollection_UpdateOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	t.Run("Upsert = true", func(t *testing.T) {
		doc := types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("version"), types.NewInt64(0),
		)

		ok, err := collection.UpdateOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(true),
		}))
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Upsert = false", func(t *testing.T) {
		doc := types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("version"), types.NewInt64(0),
		)

		ok, err := collection.UpdateOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.False(t, ok)

		_, _ = collection.InsertOne(ctx, doc)

		ok, err = collection.UpdateOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestCollection_UpdateMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	t.Run("Upsert = true", func(t *testing.T) {
		id := types.NewBinary(uuid.Must(uuid.NewV7()).Bytes())

		count, err := collection.UpdateMany(ctx, database.Where("id").Equal(id), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(true),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Upsert = false", func(t *testing.T) {
		var ids []types.Object
		for i := 0; i < batchSize; i++ {
			ids = append(ids, types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()))
		}

		var docs []types.Map
		for _, id := range ids {
			docs = append(docs, types.NewMap(
				types.NewString("id"), id,
				types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
				types.NewString("version"), types.NewInt64(0),
				types.NewString("deleted"), types.False,
			))
		}

		count, err := collection.UpdateMany(ctx, database.Where("id").In(ids...), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		_, _ = collection.InsertMany(ctx, docs)

		count, err = collection.UpdateMany(ctx, database.Where("id").In(ids...), types.NewMap(types.NewString("version"), types.NewInt64(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, len(ids), count)
	})
}

func TestCollection_DeleteOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	doc := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
	)

	ok, err := collection.DeleteOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = collection.InsertOne(ctx, doc)

	ok, err = collection.DeleteOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = collection.DeleteOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCollection_DeleteMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	var ids []types.Object
	for i := 0; i < batchSize; i++ {
		ids = append(ids, types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()))
	}

	var docs []types.Map
	for _, id := range ids {
		docs = append(docs, types.NewMap(
			types.NewString("id"), id,
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("version"), types.NewInt64(0),
			types.NewString("deleted"), types.False,
		))
	}

	count, err := collection.DeleteMany(ctx, database.Where("id").In(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = collection.InsertMany(ctx, docs)

	count, err = collection.DeleteMany(ctx, database.Where("id").In(ids...))
	assert.NoError(t, err)
	assert.Equal(t, len(ids), count)

	count, err = collection.DeleteMany(ctx, database.Where("id").In(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCollection_FindOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	doc := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("version"), types.NewInt64(0),
		types.NewString("deleted"), types.False,
	)

	res, err := collection.FindOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Nil(t, res)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	t.Run(string(database.EQ), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.NE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("id").NotEqual(doc.GetOr(types.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.GT), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").GreaterThan(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.GTE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").GreaterThanOrEqual(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.LT), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").LessThan(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.LTE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").LessThanOrEqual(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.IN), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").In(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.NIN), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").NotIn(doc.GetOr(types.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.NULL), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").IsNull())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.NNULL), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").IsNotNull())
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.AND), func(t *testing.T) {
		res, err = collection.FindOne(ctx,
			database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)).
				And(database.Where("name").Equal(doc.GetOr(types.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})

	t.Run(string(database.OR), func(t *testing.T) {
		res, err = collection.FindOne(ctx,
			database.Where("id").Equal(doc.GetOr(types.NewString("id"), nil)).
				Or(database.Where("name").Equal(doc.GetOr(types.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(types.NewString("id"), nil), res.GetOr(types.NewString("id"), nil))
	})
}

func TestCollection_FindMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	var ids []types.Object
	for i := 0; i < batchSize; i++ {
		ids = append(ids, types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()))
	}

	var docs []types.Map
	for _, id := range ids {
		docs = append(docs, types.NewMap(
			types.NewString("id"), id,
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("version"), types.NewInt64(0),
			types.NewString("deleted"), types.False,
		))
	}

	_, _ = collection.InsertMany(ctx, docs)

	t.Run(string(database.EQ), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").Equal(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.NE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").NotEqual(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.GT), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GreaterThan(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.GTE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GreaterThanOrEqual(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.LT), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").LessThan(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.LTE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").LessThanOrEqual(types.NewInt64(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.IN), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").In(ids...))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.NIN), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").NotIn(ids...))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.NULL), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").IsNull())
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.NNULL), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").IsNotNull())
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.AND), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GreaterThan(types.NewInt64(0)).And(database.Where("version").LessThanOrEqual(types.NewInt64(0))))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.OR), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GreaterThan(types.NewInt64(0)).Or(database.Where("version").LessThanOrEqual(types.NewInt64(0))))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run("Limit", func(t *testing.T) {
		limit := len(ids) / 2

		res, err := collection.FindMany(ctx, database.Where("id").In(ids...), &database.FindOptions{
			Limit: &limit,
		})
		assert.NoError(t, err)
		assert.Len(t, res, limit)
	})

	t.Run("Skip", func(t *testing.T) {
		skip := len(ids) / 2

		res, err := collection.FindMany(ctx, database.Where("id").In(ids...), &database.FindOptions{
			Skip: &skip,
		})
		assert.NoError(t, err)
		assert.Len(t, res, len(ids)-skip)
	})

	t.Run("Sorts", func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").In(ids...), &database.FindOptions{
			Sorts: []database.Sort{{Key: "id", Order: database.OrderASC}},
		})
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))

		var preID types.Object
		for _, doc := range res {
			curID := doc.GetOr(types.NewString("id"), nil)
			assert.False(t, types.Equal(preID, curID))
			preID = curID
		}
	})
}

func TestCollection_Drop(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
	))
	assert.NoError(t, err)

	err = collection.Drop(ctx)
	assert.NoError(t, err)
}

func BenchmarkCollection_InsertOne(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_InsertMany(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var docs []types.Map
		for i := 0; i < benchSize; i++ {
			docs = append(docs, types.NewMap(
				types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
				types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
			))
		}

		_, err := coll.InsertMany(ctx, docs)
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_UpdateOne(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		))
	}

	name := types.NewString(faker.UUIDHyphenated())

	v := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("name"), name,
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		next := types.NewString(faker.UUIDHyphenated())

		_, err := coll.UpdateOne(ctx, database.Where("name").Equal(name), types.NewMap(types.NewString("name"), next))
		assert.NoError(b, err)

		name = next
	}
}

func BenchmarkCollection_UpdateMany(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	name := types.NewString(faker.UUIDHyphenated())

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), name,
		))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		next := types.NewString(faker.UUIDHyphenated())

		_, err := coll.UpdateMany(ctx, database.Where("name").Equal(name), types.NewMap(types.NewString("name"), next))
		assert.NoError(b, err)

		name = next
	}
}

func BenchmarkCollection_DeleteOne(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		))
	}

	v := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err := coll.InsertOne(ctx, v)
		assert.NoError(b, err)

		b.StartTimer()

		_, err = coll.DeleteOne(ctx, database.Where("id").Equal(v.GetOr(types.NewString("id"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_DeleteMany(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	name := types.NewString(faker.UUIDHyphenated())

	var docs []types.Map
	for i := 0; i < benchSize; i++ {
		docs = append(docs, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), name,
		))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err := coll.InsertMany(ctx, docs)
		assert.NoError(b, err)

		b.StartTimer()

		_, err = coll.DeleteMany(ctx, database.Where("name").Equal(name))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_FindOne(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		))
	}

	v := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.ResetTimer()

	b.Run("With Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindOne(ctx, database.Where("id").Equal(v.GetOr(types.NewString("id"), nil)))
			assert.NoError(b, err)
		}
	})

	b.Run("Without Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindOne(ctx, database.Where("name").Equal(v.GetOr(types.NewString("name"), nil)))
			assert.NoError(b, err)
		}
	})
}

func BenchmarkCollection_FindMany(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, types.NewMap(
			types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
			types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
		))
	}

	v := types.NewMap(
		types.NewString("id"), types.NewBinary(uuid.Must(uuid.NewV7()).Bytes()),
		types.NewString("name"), types.NewString(faker.UUIDHyphenated()),
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.ResetTimer()

	b.Run("With Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindMany(ctx, database.Where("id").Equal(v.GetOr(types.NewString("id"), nil)))
			assert.NoError(b, err)
		}
	})

	b.Run("Without Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindMany(ctx, database.Where("name").Equal(v.GetOr(types.NewString("name"), nil)))
			assert.NoError(b, err)
		}
	})
}

package databasetest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
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

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
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

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("version"), primitive.NewInt(0),
	)

	_, _ = collection.InsertOne(ctx, doc)
	_, _ = collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)))
	_, _ = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
}

func TestCollection_InsertOne(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		)

		id, err := collection.InsertOne(ctx, doc)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), id)
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		)

		_, _ = collection.InsertOne(ctx, doc)

		_, err := collection.InsertOne(ctx, doc)
		assert.Error(t, err)
	})
}

func TestCollection_InsertMany(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		var docs []*primitive.Map
		for i := 0; i < batchSize; i++ {
			docs = append(docs, primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			))
		}

		ids, err := collection.InsertMany(ctx, docs)
		assert.NoError(t, err)
		assert.Len(t, ids, len(docs))
		for i, doc := range docs {
			assert.Equal(t, ids[i], doc.GetOr(primitive.NewString("id"), nil))
		}
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		var docs []*primitive.Map
		for i := 0; i < batchSize; i++ {
			docs = append(docs, primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			))
		}

		_, _ = collection.InsertMany(ctx, docs)

		_, err := collection.InsertMany(ctx, docs)
		assert.Error(t, err)
	})
}

func TestCollection_UpdateOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	t.Run("Upsert = true", func(t *testing.T) {
		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("version"), primitive.NewInt(0),
		)

		ok, err := collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(true),
		}))
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Upsert = false", func(t *testing.T) {
		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("version"), primitive.NewInt(0),
		)

		ok, err := collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.False(t, ok)

		_, _ = collection.InsertOne(ctx, doc)

		ok, err = collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestCollection_UpdateMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	t.Run("Upsert = true", func(t *testing.T) {
		id := primitive.NewBinary(ulid.Make().Bytes())

		count, err := collection.UpdateMany(ctx, database.Where("id").EQ(id), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(true),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Upsert = false", func(t *testing.T) {
		var ids []primitive.Value
		for i := 0; i < batchSize; i++ {
			ids = append(ids, primitive.NewBinary(ulid.Make().Bytes()))
		}

		var docs []*primitive.Map
		for _, id := range ids {
			docs = append(docs, primitive.NewMap(
				primitive.NewString("id"), id,
				primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			))
		}

		count, err := collection.UpdateMany(ctx, database.Where("id").IN(ids...), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		_, _ = collection.InsertMany(ctx, docs)

		count, err = collection.UpdateMany(ctx, database.Where("id").IN(ids...), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, len(ids), count)
	})
}

func TestCollection_DeleteOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
	)

	ok, err := collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, _ = collection.InsertOne(ctx, doc)

	ok, err = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCollection_DeleteMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	var ids []primitive.Value
	for i := 0; i < batchSize; i++ {
		ids = append(ids, primitive.NewBinary(ulid.Make().Bytes()))
	}

	var docs []*primitive.Map
	for _, id := range ids {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), id,
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		))
	}

	count, err := collection.DeleteMany(ctx, database.Where("id").IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, _ = collection.InsertMany(ctx, docs)

	count, err = collection.DeleteMany(ctx, database.Where("id").IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, len(ids), count)

	count, err = collection.DeleteMany(ctx, database.Where("id").IN(ids...))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCollection_FindOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		primitive.NewString("version"), primitive.NewInt(0),
		primitive.NewString("deleted"), primitive.FALSE,
	)

	res, err := collection.FindOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Nil(t, res)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	t.Run(string(database.EQ), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.NE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("id").NE(doc.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.GT), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").GT(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.GTE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").GTE(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.LT), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").LT(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run(string(database.LTE), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").LTE(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.IN), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").IN(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.NIN), func(t *testing.T) {
		res, err = collection.FindOne(ctx, database.Where("version").NotIN(doc.GetOr(primitive.NewString("version"), nil)))
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
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.AND), func(t *testing.T) {
		res, err = collection.FindOne(ctx,
			database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)).
				And(database.Where("name").EQ(doc.GetOr(primitive.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})

	t.Run(string(database.OR), func(t *testing.T) {
		res, err = collection.FindOne(ctx,
			database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)).
				Or(database.Where("name").EQ(doc.GetOr(primitive.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), res.GetOr(primitive.NewString("id"), nil))
	})
}

func TestCollection_FindMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	var ids []primitive.Value
	for i := 0; i < batchSize; i++ {
		ids = append(ids, primitive.NewBinary(ulid.Make().Bytes()))
	}

	var docs []*primitive.Map
	for _, id := range ids {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), id,
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		))
	}

	_, _ = collection.InsertMany(ctx, docs)

	t.Run(string(database.EQ), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").EQ(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.NE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").NE(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.GT), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GT(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.GTE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GTE(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.LT), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").LT(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.LTE), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").LTE(primitive.NewInt(0)))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.IN), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").IN(ids...))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run(string(database.NIN), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").NotIN(ids...))
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
		res, err := collection.FindMany(ctx, database.Where("version").GT(primitive.NewInt(0)).And(database.Where("version").LTE(primitive.NewInt(0))))
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(string(database.OR), func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("version").GT(primitive.NewInt(0)).Or(database.Where("version").LTE(primitive.NewInt(0))))
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))
	})

	t.Run("Limit", func(t *testing.T) {
		limit := len(ids) / 2

		res, err := collection.FindMany(ctx, database.Where("id").IN(ids...), &database.FindOptions{
			Limit: &limit,
		})
		assert.NoError(t, err)
		assert.Len(t, res, limit)
	})

	t.Run("Skip", func(t *testing.T) {
		skip := len(ids) / 2

		res, err := collection.FindMany(ctx, database.Where("id").IN(ids...), &database.FindOptions{
			Skip: &skip,
		})
		assert.NoError(t, err)
		assert.Len(t, res, len(ids)-skip)
	})

	t.Run("Sorts", func(t *testing.T) {
		res, err := collection.FindMany(ctx, database.Where("id").IN(ids...), &database.FindOptions{
			Sorts: []database.Sort{{Key: "id", Order: database.OrderASC}},
		})
		assert.NoError(t, err)
		assert.Len(t, res, len(ids))

		var preID primitive.Value
		for _, doc := range res {
			curID := doc.GetOr(primitive.NewString("id"), nil)
			assert.Equal(t, primitive.Compare(preID, curID), -1)
			preID = curID
		}
	})
}

func TestCollection_Drop(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
	))
	assert.NoError(t, err)

	err = collection.Drop(ctx)
	assert.NoError(t, err)
}

func BenchmarkCollection_InsertOne(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < b.N; i++ {
		_, err := coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_InsertMany(b *testing.B, coll database.Collection) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < b.N; i++ {
		var docs []*primitive.Map
		for i := 0; i < benchSize; i++ {
			docs = append(docs, primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
			))
		}

		_, err := coll.InsertMany(ctx, docs)
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_UpdateOne(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		))
	}

	name := primitive.NewString(faker.UUIDHyphenated())

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), name,
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		next := primitive.NewString(faker.UUIDHyphenated())

		_, err := coll.UpdateOne(ctx, database.Where("name").EQ(name), primitive.NewMap(primitive.NewString("name"), next))
		assert.NoError(b, err)

		name = next
	}
}

func BenchmarkCollection_UpdateMany(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	name := primitive.NewString(faker.UUIDHyphenated())

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), name,
		))
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		next := primitive.NewString(faker.UUIDHyphenated())

		_, err := coll.UpdateMany(ctx, database.Where("name").EQ(name), primitive.NewMap(primitive.NewString("name"), next))
		assert.NoError(b, err)

		name = next
	}
}

func BenchmarkCollection_DeleteOne(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
	)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err := coll.InsertOne(ctx, v)
		assert.NoError(b, err)

		b.StartTimer()

		_, err = coll.DeleteOne(ctx, database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_DeleteMany(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	name := primitive.NewString(faker.UUIDHyphenated())

	var docs []*primitive.Map
	for i := 0; i < benchSize; i++ {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), name,
		))
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err := coll.InsertMany(ctx, docs)
		assert.NoError(b, err)

		b.StartTimer()

		_, err = coll.DeleteMany(ctx, database.Where("name").EQ(name))
		assert.NoError(b, err)
	}
}

func BenchmarkCollection_FindOne(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.StartTimer()

	b.Run("With Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindOne(ctx, database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
			assert.NoError(b, err)
		}
	})

	b.Run("Without Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindOne(ctx, database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)))
			assert.NoError(b, err)
		}
	})
}

func BenchmarkCollection_FindMany(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < benchSize; i++ {
		_, _ = coll.InsertOne(ctx, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.UUIDHyphenated()),
	)

	_, err := coll.InsertOne(ctx, v)
	assert.NoError(b, err)

	b.StartTimer()

	b.Run("With Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindMany(ctx, database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
			assert.NoError(b, err)
		}
	})

	b.Run("Without Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.FindMany(ctx, database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)))
			assert.NoError(b, err)
		}
	})
}

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

const (
	benchmarkSetSize = 1000
)

func AssertCollectionName(t *testing.T, collection database.Collection) {
	t.Helper()

	name := collection.Name()
	assert.NotEmpty(t, name)
}

func AssertCollectionIndexes(t *testing.T, collection database.Collection) {
	t.Helper()

	indexes := collection.Indexes()
	assert.NotNil(t, indexes)
}

func AssertCollectionWatch(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := collection.Watch(ctx, nil)
	assert.NoError(t, err)
	defer stream.Close()

	go func() {
		for {
			event, ok := <-stream.Next()
			if ok {
				assert.NotNil(t, event.DocumentID)
			} else {
				return
			}
		}
	}()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("version"), primitive.NewInt(0),
	)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	_, err = collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)))
	assert.NoError(t, err)

	_, err = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
}

func AssertCollectionInsert(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		)

		_, _ = collection.InsertOne(ctx, doc)

		_, err := collection.InsertOne(ctx, doc)
		assert.Error(t, err)
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
			primitive.NewString("version"), primitive.NewInt(0),
			primitive.NewString("deleted"), primitive.FALSE,
		)

		id, err := collection.InsertOne(ctx, doc)
		assert.NoError(t, err)
		assert.Equal(t, doc.GetOr(primitive.NewString("id"), nil), id)
	})
}

func AssertCollectionInsertMany(t *testing.T, collection database.Collection) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		docs := []*primitive.Map{
			primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.Word()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			),
			primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.Word()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			),
		}

		ids, err := collection.InsertMany(ctx, docs)
		assert.NoError(t, err)
		assert.Len(t, ids, len(docs))
		for i, doc := range docs {
			assert.Equal(t, ids[i], doc.GetOr(primitive.NewString("id"), nil))
		}
	})

	t.Run("Conflict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		docs := []*primitive.Map{
			primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.Word()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			),
			primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.Word()),
				primitive.NewString("version"), primitive.NewInt(0),
				primitive.NewString("deleted"), primitive.FALSE,
			),
		}

		_, _ = collection.InsertMany(ctx, docs)

		_, err := collection.InsertMany(ctx, docs)
		assert.Error(t, err)
	})
}

func AssertCollectionUpdateOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("options.Upsert = true", func(t *testing.T) {
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

	t.Run("options.Upsert = false", func(t *testing.T) {
		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("version"), primitive.NewInt(0),
		)

		ok, err := collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.False(t, ok)

		_, err = collection.InsertOne(ctx, doc)
		assert.NoError(t, err)

		ok, err = collection.UpdateOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func AssertCollectionUpdateMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("options.Upsert = true", func(t *testing.T) {
		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("version"), primitive.NewInt(0),
		)

		count, err := collection.UpdateMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(true),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("options.Upsert = false", func(t *testing.T) {
		doc := primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("version"), primitive.NewInt(0),
		)

		count, err := collection.UpdateMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		_, err = collection.InsertOne(ctx, doc)
		assert.NoError(t, err)

		count, err = collection.UpdateMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(primitive.NewString("version"), primitive.NewInt(1)), lo.ToPtr(database.UpdateOptions{
			Upsert: lo.ToPtr(false),
		}))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func AssertCollectionDeleteOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
	)

	ok, err := collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	ok, err = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = collection.DeleteOne(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func AssertCollectionDeleteMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
	)

	count, err := collection.DeleteMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	count, err = collection.DeleteMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = collection.DeleteMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func AssertCollectionFindOne(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
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

func AssertCollectionFindMany(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	doc := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
		primitive.NewString("version"), primitive.NewInt(0),
		primitive.NewString("deleted"), primitive.FALSE,
	)

	res, err := collection.FindMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
	assert.NoError(t, err)
	assert.Len(t, res, 0)

	_, err = collection.InsertOne(ctx, doc)
	assert.NoError(t, err)

	t.Run(string(database.EQ), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.NE), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("id").NE(doc.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run(string(database.GT), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").GT(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run(string(database.GTE), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").GTE(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.LT), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").LT(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run(string(database.LTE), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").LTE(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.IN), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").IN(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.NIN), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").NotIN(doc.GetOr(primitive.NewString("version"), nil)))
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run(string(database.NULL), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").IsNull())
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run(string(database.NNULL), func(t *testing.T) {
		res, err = collection.FindMany(ctx, database.Where("version").IsNotNull())
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.AND), func(t *testing.T) {
		res, err = collection.FindMany(ctx,
			database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)).
				And(database.Where("name").EQ(doc.GetOr(primitive.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run(string(database.OR), func(t *testing.T) {
		res, err = collection.FindMany(ctx,
			database.Where("id").EQ(doc.GetOr(primitive.NewString("id"), nil)).
				Or(database.Where("name").EQ(doc.GetOr(primitive.NewString("name"), nil))),
		)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}

func AssertCollectionDrop(t *testing.T, collection database.Collection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
	))
	assert.NoError(t, err)

	err = collection.Drop(ctx)
	assert.NoError(t, err)
}

func BenchmarkCollectionInsertOne(b *testing.B, coll database.Collection) {
	b.Helper()

	for i := 0; i < b.N; i++ {
		_, err := coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionInsertMany(b *testing.B, coll database.Collection) {
	b.Helper()

	for i := 0; i < b.N; i++ {
		var docs []*primitive.Map
		for j := 0; j < 10; j++ {
			docs = append(docs, primitive.NewMap(
				primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
				primitive.NewString("name"), primitive.NewString(faker.Word()),
			))
		}

		_, err := coll.InsertMany(context.Background(), docs)
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionUpdateOne(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	_, err := coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.UpdateOne(context.Background(), database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)), primitive.NewMap(
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionUpdateMany(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	var docs []*primitive.Map
	for j := 0; j < 10; j++ {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), v.GetOr(primitive.NewString("name"), nil),
		))
	}
	_, err := coll.InsertMany(context.Background(), docs)
	assert.NoError(b, err)

	_, err = coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.UpdateMany(context.Background(), database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)), primitive.NewMap(
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionDeleteOne(b *testing.B, coll database.Collection) {
	b.Helper()

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_, err := coll.InsertOne(context.Background(), v)
		assert.NoError(b, err)
		b.StartTimer()

		_, err = coll.DeleteOne(context.Background(), database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionDeleteMany(b *testing.B, coll database.Collection) {
	b.Helper()

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	var docs []*primitive.Map
	for j := 0; j < 10; j++ {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), v.GetOr(primitive.NewString("name"), nil),
		))
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err := coll.InsertMany(context.Background(), docs)
		assert.NoError(b, err)

		_, err = coll.InsertOne(context.Background(), v)
		assert.NoError(b, err)

		b.StartTimer()

		_, err = coll.DeleteMany(context.Background(), database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionFindOneWithIndex(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	_, err := coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.FindOne(context.Background(), database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionFindOneWithoutIndex(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	_, err := coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.FindOne(context.Background(), database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionFindManyWithIndex(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	var docs []*primitive.Map
	for j := 0; j < 10; j++ {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), v.GetOr(primitive.NewString("name"), nil),
		))
	}

	_, err := coll.InsertMany(context.Background(), docs)
	assert.NoError(b, err)

	_, err = coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.FindMany(context.Background(), database.Where("id").EQ(v.GetOr(primitive.NewString("id"), nil)))
		assert.NoError(b, err)
	}
}

func BenchmarkCollectionFindManyWithoutIndex(b *testing.B, coll database.Collection) {
	b.Helper()
	b.StopTimer()

	for i := 0; i < benchmarkSetSize; i++ {
		_, _ = coll.InsertOne(context.Background(), primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), primitive.NewString(faker.Word()),
		))
	}

	v := primitive.NewMap(
		primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
		primitive.NewString("name"), primitive.NewString(faker.Word()),
	)

	var docs []*primitive.Map
	for j := 0; j < 10; j++ {
		docs = append(docs, primitive.NewMap(
			primitive.NewString("id"), primitive.NewBinary(ulid.Make().Bytes()),
			primitive.NewString("name"), v.GetOr(primitive.NewString("name"), nil),
		))
	}

	_, err := coll.InsertMany(context.Background(), docs)
	assert.NoError(b, err)

	_, err = coll.InsertOne(context.Background(), v)
	assert.NoError(b, err)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := coll.FindMany(context.Background(), database.Where("name").EQ(v.GetOr(primitive.NewString("name"), nil)))
		assert.NoError(b, err)
	}
}

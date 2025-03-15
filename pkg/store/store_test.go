package store

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := New()

	strm, err := s.Watch(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, strm)

	defer strm.Close()

	var count atomic.Int32
	go func() {
		for range strm.Next() {
			count.Add(1)
		}
	}()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err = s.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return count.Load() == 1
	}, 1*time.Second, 10*time.Millisecond)

	c, err := s.Delete(ctx, Where(KeyID).Equal(doc.Get(types.NewString(KeyID))))
	require.NoError(t, err)
	require.Equal(t, 1, c)
	require.Eventually(t, func() bool {
		return count.Load() == 2
	}, 1*time.Second, 10*time.Millisecond)
}

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)

	err = s.Index(ctx, []types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestStore_Unindex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)

	err = s.Index(ctx, []types.String{types.NewString("name")})
	require.NoError(t, err)

	err = s.Unindex(ctx, []types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestStore_Insert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
		types.NewString("age"), types.NewInt(123),
	)

	err := s.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)
}

func TestStore_Update(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		count, err := s.Update(
			ctx,
			Where(KeyID).Equal(doc.Get(types.NewString(KeyID))),
			Set(types.NewMap(types.NewString("name"), types.NewString(faker.Word()))),
		)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("{'upsert': true}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		count, err := s.Update(
			ctx,
			Where(KeyID).Equal(types.NewString(faker.UUIDHyphenated())),
			Set(types.NewMap(types.NewString("name"), types.NewString(faker.Word()))),
			UpdateOptions{Upsert: true},
		)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}

func TestStore_Delete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)

	count, err := s.Delete(ctx, Where(KeyID).Equal(doc.Get(types.NewString(KeyID))))
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStore_Find(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, nil)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("nil, {'limit': <limit>, 'sort': <sort>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(150),
		)
		doc2 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(50),
		)

		err := s.Insert(ctx, []types.Map{doc1, doc2})
		require.NoError(t, err)

		docs, err := s.Find(ctx, nil, FindOptions{
			Limit: 1,
			Sort:  types.NewMap(types.NewString(KeyID), types.NewInt(1)),
		})
		require.NoError(t, err)
		require.Len(t, docs, 1)
	})

	t.Run("{'id': {'$eq': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where(KeyID).Equal(doc.Get(types.NewString(KeyID))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'id': {'$ne': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where(KeyID).NotEqual(types.NewString(faker.UUIDHyphenated())))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$gt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err = s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where("age").GreaterThan(types.NewInt(0)))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$gte': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err = s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where("age").GreaterThanOrEqual(types.NewInt(0)))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$lt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err = s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where("age").LessThan(types.NewInt(321)))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$lte': <name>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err = s.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Where("age").LessThanOrEqual(types.NewInt(321)))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{$and: [{'age': {'$gt': 100}}, {'name': {'$eq': <name>}}]}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc1 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)
		doc2 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(50),
		)

		err = s.Insert(ctx, []types.Map{doc1, doc2})
		require.NoError(t, err)

		docs, err := s.Find(ctx, And(Where("age").GreaterThan(types.NewInt(0)), Where("name").Equal(doc1.Get(types.NewString("name")))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc1, docs[0])
	})

	t.Run("{$or: [{'age': {'$lt': 100}}, {'name': {'$eq': <name>}}]}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []types.String{types.NewString("name")})
		require.NoError(t, err)
		err = s.Index(ctx, []types.String{types.NewString("age")})
		require.NoError(t, err)

		doc1 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(150),
		)
		doc2 := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(50),
		)

		err = s.Insert(ctx, []types.Map{doc1, doc2})
		require.NoError(t, err)

		docs, err := s.Find(ctx, Or(Where("age").LessThan(types.NewInt(100)), Where("name").Equal(doc1.Get(types.NewString("name")))))
		require.NoError(t, err)
		require.Len(t, docs, 2)
	})
}

func BenchmarkStore_Insert(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	for i := 0; i < b.N; i++ {
		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(b, err)
	}
}

func BenchmarkStore_Update(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, []types.Map{doc})
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		count, err := s.Update(
			ctx,
			Where(KeyID).Equal(doc.Get(types.NewString(KeyID))),
			Set(types.NewMap(types.NewString("name"), types.NewString(faker.Word()))),
		)
		require.NoError(b, err)
		require.Equal(b, 1, count)
	}
}

func BenchmarkStore_Delete(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		doc := types.NewMap(
			types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)

		err := s.Insert(ctx, []types.Map{doc})
		require.NoError(b, err)

		b.StartTimer()

		count, err := s.Delete(ctx, Where(KeyID).Equal(doc.Get(types.NewString(KeyID))))
		require.NoError(b, err)
		require.Equal(b, 1, count)
	}
}

func BenchmarkStore_Find(b *testing.B) {
	b.Run("with index", func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		keys := make([]types.String, 0, b.N)

		for i := 0; i < b.N; i++ {
			doc := types.NewMap(
				types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
				types.NewString("name"), types.NewString(faker.Word()),
				types.NewString("age"), types.NewInt(123),
			)

			err := s.Insert(ctx, []types.Map{doc})
			require.NoError(b, err)

			keys = append(keys, doc.Get(types.NewString(KeyID)).(types.String))
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, err := s.Find(ctx, Where(KeyID).Equal(key))
			require.NoError(b, err)
		}
	})

	b.Run("without index", func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		keys := make([]types.String, 0, b.N)

		for i := 0; i < b.N; i++ {
			doc := types.NewMap(
				types.NewString(KeyID), types.NewString(faker.UUIDHyphenated()),
				types.NewString("name"), types.NewString(faker.Word()),
				types.NewString("age"), types.NewInt(123),
			)

			err := s.Insert(ctx, []types.Map{doc})
			require.NoError(b, err)

			keys = append(keys, doc.Get(types.NewString("name")).(types.String))
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, err := s.Find(ctx, Where("name").Equal(key))
			require.NoError(b, err)
		}
	})
}

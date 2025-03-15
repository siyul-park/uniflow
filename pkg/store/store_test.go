package store

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, doc)
	require.NoError(t, err)

	err = s.Index(ctx, []types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestStore_Unindex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Insert(ctx, doc)
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
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
		types.NewString("age"), types.NewInt(123),
	)

	err := s.Insert(ctx, doc)
	require.NoError(t, err)
}

func TestStore_Remove(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Insert(ctx, doc)

	count, err := s.Remove(ctx, types.NewMap(primaryKey, doc.Get(primaryKey)))
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStore_Find(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, nil)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'id': <id>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(primaryKey, doc.Get(primaryKey)))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'id': {'$eq': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(primaryKey, types.NewMap(types.NewString("$eq"), doc.Get(primaryKey))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'id': {'$ne': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(primaryKey, types.NewMap(types.NewString("$ne"), types.NewString(faker.UUIDHyphenated()))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$gt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$gt"), types.NewInt(0))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$gte': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$gte"), types.NewInt(0))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$lt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$lt"), types.NewInt(321))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{'age': {'$lte': <name>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$lte"), types.NewInt(321))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("{$and: [ {'age': {'$gt': 100}}, {'name': {'$eq': <name>}} ] }", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()
		name := types.NewString(faker.Word())

		doc1 := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), name,
			types.NewString("age"), types.NewInt(123),
		)
		doc2 := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(50),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc1, doc2)

		query := types.NewMap(
			types.NewString("$and"), types.NewSlice(
				types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$gt"), types.NewInt(100))),
				types.NewMap(types.NewString("name"), types.NewMap(types.NewString("$eq"), name)),
			),
		)
		docs, err := s.Find(ctx, query)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc1, docs[0])
	})

	t.Run("{$or: [ {'age': {'$lt': 100}}, {'name': {'$eq': <name>}} ] }", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()
		name := types.NewString(faker.Word())

		doc1 := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), name,
			types.NewString("age"), types.NewInt(150),
		)
		doc2 := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(50),
		)

		_ = s.Index(ctx, []types.String{types.NewString("name")})
		_ = s.Index(ctx, []types.String{types.NewString("age")})

		_ = s.Insert(ctx, doc1, doc2)

		query := types.NewMap(
			types.NewString("$or"), types.NewSlice(
				types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$lt"), types.NewInt(100))),
				types.NewMap(types.NewString("name"), types.NewMap(types.NewString("$eq"), name)),
			),
		)
		docs, err := s.Find(ctx, query)
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
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		err := s.Insert(ctx, doc)
		require.NoError(b, err)
	}
}

func BenchmarkStore_Remove(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)

		_ = s.Insert(ctx, doc)

		b.StartTimer()

		count, err := s.Remove(ctx, types.NewMap(primaryKey, doc.Get(primaryKey)))
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
				primaryKey, types.NewString(faker.UUIDHyphenated()),
				types.NewString("name"), types.NewString(faker.Word()),
				types.NewString("age"), types.NewInt(123),
			)

			_ = s.Insert(ctx, doc)
			keys = append(keys, doc.Get(primaryKey).(types.String))
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, err := s.Find(ctx, types.NewMap(primaryKey, key))
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
				primaryKey, types.NewString(faker.UUIDHyphenated()),
				types.NewString("name"), types.NewString(faker.Word()),
				types.NewString("age"), types.NewInt(123),
			)

			_ = s.Insert(ctx, doc)
			keys = append(keys, doc.Get(types.NewString("name")).(types.String))
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, err := s.Find(ctx, types.NewMap(types.NewString("name"), key))
			require.NoError(b, err)
		}
	})
}

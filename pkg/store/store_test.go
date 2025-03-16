package store

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := New()

	strm, err := s.Watch(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, strm)

	defer strm.Close(ctx)

	var count atomic.Int32
	go func() {
		for strm.Next(ctx) {
			count.Add(1)
		}
	}()

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err = s.Insert(ctx, []any{doc})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return count.Load() == 1 }, time.Second, 10*time.Millisecond)

	_, err = s.Delete(ctx, map[string]any{"id": doc["id"]})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return count.Load() == 2 }, time.Second, 10*time.Millisecond)
}

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)

	err = s.Index(ctx, []string{"name"}, IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": 1}},
	})
	require.NoError(t, err)

	err = s.Index(ctx, []string{"name"})
	require.NoError(t, err)
}

func TestStore_Unindex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)

	err = s.Index(ctx, []string{"name"})
	require.NoError(t, err)

	err = s.Unindex(ctx, []string{"name"})
	require.NoError(t, err)

	err = s.Unindex(ctx, []string{"name"})
	require.NoError(t, err)
}

func TestStore_Insert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := New()

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)
}

func TestStore_Update(t *testing.T) {
	t.Run("{'$set': <doc>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}

		err := s.Insert(ctx, []any{doc})
		require.NoError(t, err)

		count, err := s.Update(
			ctx,
			map[string]any{"id": doc["id"]},
			map[string]any{"$set": map[string]any{"name": faker.Name()}},
		)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("{'$unset': <doc>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}

		err := s.Insert(ctx, []any{doc})
		require.NoError(t, err)

		count, err := s.Update(
			ctx,
			map[string]any{"id": doc["id"]},
			map[string]any{"$unset": map[string]any{"name": nil}},
		)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("{'$set': <doc>}, {'upsert': true}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}

		err := s.Insert(ctx, []any{doc})
		require.NoError(t, err)

		count, err := s.Update(
			ctx,
			map[string]any{"$or": []map[string]any{{"id": faker.UUIDHyphenated()}, {"name": faker.UUIDHyphenated()}}},
			map[string]any{"$set": map[string]any{"name": faker.Name()}},
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

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)

	count, err := s.Delete(ctx, map[string]any{"id": doc["id"]})
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStore_Find(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, nil)
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 2)
	})

	t.Run("{limit: 1, sort: {'id': 1}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, nil, FindOptions{Limit: 1, Sort: map[string]any{"id": 1}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'id': <id>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"id": doc1["id"]})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'id': {'$exists': <exists>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"id": map[string]any{"$exists": 1}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 2)
	})

	t.Run("{'id': {'$eq': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"id": map[string]any{"$eq": doc1["id"]}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'id': {'$ne': <id>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"id": map[string]any{"$ne": doc1["id"]}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'version': {'$gt': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []string{"version"})
		require.NoError(t, err)

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err = s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"version": map[string]any{"$gt": 1}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'version': {'$gte': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []string{"version"})
		require.NoError(t, err)

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err = s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"version": map[string]any{"$gte": 1}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 2)
	})

	t.Run("{'version': {'$lt': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []string{"version"})
		require.NoError(t, err)

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err = s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"version": map[string]any{"$lt": 2}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'version': {'$lte': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		err := s.Index(ctx, []string{"version"})
		require.NoError(t, err)

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err = s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"version": map[string]any{"$lte": 2}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 2)
	})

	t.Run("{'$and': [{'id': {'$eq': <id>}}]}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"$and": []any{map[string]any{"id": map[string]any{"$eq": doc1["id"]}}, map[string]any{"id": map[string]any{"$eq": doc2["id"]}}}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 0)
	})

	t.Run("{'$or': [{'id': {'$eq': <id>}}]}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc1 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		doc2 := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 2,
		}

		err := s.Insert(ctx, []any{doc1, doc2})
		require.NoError(t, err)

		c, err := s.Find(ctx, map[string]any{"$or": []any{map[string]any{"id": map[string]any{"$eq": doc1["id"]}}, map[string]any{"id": map[string]any{"$eq": doc2["id"]}}}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 2)
	})
}

func BenchmarkStore_Insert(b *testing.B) {
	ctx := context.TODO()

	s := New()

	for i := 0; i < b.N; i++ {
		doc := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		require.NoError(b, s.Insert(ctx, []any{doc}))
	}
}

func BenchmarkStore_Find(b *testing.B) {
	ctx := context.TODO()

	s := New()

	docs := make([]map[string]any, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		require.NoError(b, s.Insert(ctx, []any{docs[i]}))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c, err := s.Find(ctx, map[string]any{"id": docs[i]["id"]})
		require.NoError(b, err)

		for c.Next(ctx) {
			var doc map[string]any
			require.NoError(b, c.Decode(&doc))
		}
		require.NoError(b, c.Close(ctx))
	}
}

func BenchmarkStore_Update(b *testing.B) {
	ctx := context.TODO()

	s := New()

	docs := make([]map[string]any, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		require.NoError(b, s.Insert(ctx, []any{docs[i]}))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count, err := s.Update(ctx, map[string]any{"id": docs[i]["id"]}, map[string]any{"$set": map[string]any{"version": i}})
		require.NoError(b, err)
		require.Equal(b, 1, count)
	}
}

func BenchmarkStore_Delete(b *testing.B) {
	ctx := context.TODO()

	s := New()

	docs := make([]map[string]any, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": 1,
		}
		require.NoError(b, s.Insert(ctx, []any{docs[i]}))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count, err := s.Delete(ctx, map[string]any{"id": docs[i]["id"]})
		require.NoError(b, err)
		require.Equal(b, 1, count)
	}
}

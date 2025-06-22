package driver

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/siyul-park/uniflow/plugins/mongodb/internal/server"
)

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	err := s.Index(ctx, []string{"name"}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": 1}},
	})
	require.NoError(t, err)

	indexes, err := s.Indexes(ctx)
	require.NoError(t, err)
	require.Len(t, indexes, 2)

	err = s.Index(ctx, []string{"name"})
	require.NoError(t, err)
}

func TestStore_Unindex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	err := s.Index(ctx, []string{"name"})
	require.NoError(t, err)

	err = s.Unindex(ctx, []string{"name"})
	require.NoError(t, err)

	err = s.Unindex(ctx, []string{"name"})
	require.NoError(t, err)
}

func TestStore_Insert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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
			driver.UpdateOptions{Upsert: true},
		)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}

func TestStore_Delete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		c, err := s.Find(ctx, nil, driver.FindOptions{Limit: 1, Sort: map[string]any{"id": 1}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'id': <id>}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		c, err := s.Find(ctx, map[string]any{"id": map[string]any{"$ne": doc1["id"]}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'version': {'$gt': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		c, err := s.Find(ctx, map[string]any{"version": map[string]any{"$lt": 2}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 1)
	})

	t.Run("{'version': {'$lte': <version>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

		c, err := s.Find(ctx, map[string]any{"$and": []any{map[string]any{"id": map[string]any{"$eq": doc1["id"]}}, map[string]any{"id": map[string]any{"$eq": doc2["id"]}}}})
		require.NoError(t, err)

		var docs []map[string]any
		require.NoError(t, c.All(ctx, &docs))
		require.Len(t, docs, 0)
	})

	t.Run("{'$or': [{'id': {'$eq': <id>}}]}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		srv := server.New()
		defer server.Release(srv)

		con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
		defer con.Disconnect(ctx)

		s := NewStore(con.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

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

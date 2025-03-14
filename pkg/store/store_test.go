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
	t.Run("query: nil", func(t *testing.T) {
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

	t.Run("query: {'id': <id>}", func(t *testing.T) {
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

	t.Run("query: {'id': {'$eq': <id>}}", func(t *testing.T) {
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

	t.Run("query: {'id': {'$ne': <id>}}", func(t *testing.T) {
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

	t.Run("query: {'age': {'$gt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$gt"), types.NewInt(0))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("query: {'age': {'$gte': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$gte"), types.NewInt(0))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("query: {'age': {'$lt': <age>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$lt"), types.NewInt(321))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})

	t.Run("query: {'name': {'$lte': <name>}}", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := New()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
			types.NewString("age"), types.NewInt(123),
		)

		_ = s.Insert(ctx, doc)

		docs, err := s.Find(ctx, types.NewMap(types.NewString("age"), types.NewMap(types.NewString("$lte"), types.NewInt(321))))
		require.NoError(t, err)
		require.Len(t, docs, 1)
		require.Equal(t, doc, docs[0])
	})
}

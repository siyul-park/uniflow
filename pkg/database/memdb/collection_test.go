package memdb

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/databasetest"
)

func TestCollection_Name(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionName(t, coll)
}

func TestCollection_Indexes(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionIndexes(t, coll)
}

func TestCollection_Watch(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionWatch(t, coll)
}

func TestCollection_InsertOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionInsertOne(t, coll)
}

func TestCollection_InsertMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionInsertMany(t, coll)
}

func TestCollection_UpdateOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionUpdateOne(t, coll)
}

func TestCollection_UpdateMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionUpdateMany(t, coll)
}

func TestCollection_DeleteOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionDeleteOne(t, coll)
}

func TestCollection_DeleteMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionDeleteMany(t, coll)
}

func TestCollection_FindOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionFindOne(t, coll)
}

func TestCollection_FindMany(t *testing.T) {
	coll := NewCollection(faker.Name())
	databasetest.AssertCollectionFindMany(t, coll)
}

func TestCollection_Drop(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.AssertCollectionDrop(t, coll)
}

func BenchmarkCollection_InsertOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionInsertOne(b, coll)
}

func BenchmarkCollection_InsertMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionInsertMany(b, coll)
}

func BenchmarkCollection_UpdateOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionUpdateOne(b, coll)
}

func BenchmarkCollection_UpdateMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionUpdateMany(b, coll)
}

func BenchmarkCollection_DeleteOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionDeleteOne(b, coll)
}

func BenchmarkCollection_DeleteMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollectionDeleteMany(b, coll)
}

func BenchmarkCollection_FindOne(b *testing.B) {
	b.Run("with index", func(b *testing.B) {
		coll := NewCollection(faker.Name())

		databasetest.BenchmarkCollectionFindOneWithIndex(b, coll)
	})

	b.Run("without index", func(b *testing.B) {
		coll := NewCollection(faker.Name())

		databasetest.BenchmarkCollectionFindOneWithoutIndex(b, coll)
	})
}

func BenchmarkCollection_FindMany(b *testing.B) {
	b.Run("with index", func(b *testing.B) {
		coll := NewCollection(faker.Name())

		databasetest.BenchmarkCollectionFindManyWithIndex(b, coll)
	})

	b.Run("without index", func(b *testing.B) {
		coll := NewCollection(faker.Name())

		databasetest.BenchmarkCollectionFindManyWithoutIndex(b, coll)
	})
}

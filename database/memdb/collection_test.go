package memdb

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database/databasetest"
)

func TestCollection_Name(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_Name(t, coll)
}

func TestCollection_Indexes(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_Indexes(t, coll)
}

func TestCollection_Watch(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_Watch(t, coll)
}

func TestCollection_InsertOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_InsertOne(t, coll)
}

func TestCollection_InsertMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_InsertMany(t, coll)
}

func TestCollection_UpdateOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_UpdateOne(t, coll)
}

func TestCollection_UpdateMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_UpdateMany(t, coll)
}

func TestCollection_DeleteOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_DeleteOne(t, coll)
}

func TestCollection_DeleteMany(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_DeleteMany(t, coll)
}

func TestCollection_FindOne(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_FindOne(t, coll)
}

func TestCollection_FindMany(t *testing.T) {
	coll := NewCollection(faker.Name())
	databasetest.TestCollection_FindMany(t, coll)
}

func TestCollection_Drop(t *testing.T) {
	coll := NewCollection(faker.Name())

	databasetest.TestCollection_Drop(t, coll)
}

func BenchmarkCollection_InsertOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_InsertOne(b, coll)
}

func BenchmarkCollection_InsertMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_InsertMany(b, coll)
}

func BenchmarkCollection_UpdateOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_UpdateOne(b, coll)
}

func BenchmarkCollection_UpdateMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_UpdateMany(b, coll)
}

func BenchmarkCollection_DeleteOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_DeleteOne(b, coll)
}

func BenchmarkCollection_DeleteMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_DeleteMany(b, coll)
}

func BenchmarkCollection_FindOne(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_FindOne(b, coll)
}

func BenchmarkCollection_FindMany(b *testing.B) {
	coll := NewCollection(faker.Name())

	databasetest.BenchmarkCollection_FindMany(b, coll)
}

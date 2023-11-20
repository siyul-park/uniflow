package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/databasetest"
	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCollection_Name(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionName(t, coll)
}

func TestCollection_Indexes(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionIndexes(t, coll)
}

func TestCollection_Watch(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionWatch(t, coll)
}

func TestCollection_InsertOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionInsertOne(t, coll)
}

func TestCollection_InsertMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionInsertMany(t, coll)
}

func TestCollection_UpdateOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionUpdateOne(t, coll)
}

func TestCollection_UpdateMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionUpdateMany(t, coll)
}

func TestCollection_DeleteOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionDeleteOne(t, coll)
}

func TestCollection_DeleteMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionDeleteMany(t, coll)
}

func TestCollection_FindOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionFindOne(t, coll)
}

func TestCollection_FindMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionFindMany(t, coll)
}

func TestCollection_Drop(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.AssertCollectionDrop(t, coll)
}

func BenchmarkCollection_InsertOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionInsertOne(b, coll)
}

func BenchmarkCollection_InsertMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionInsertMany(b, coll)
}

func BenchmarkCollection_UpdateOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionUpdateOne(b, coll)
}

func BenchmarkCollection_UpdateMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionUpdateMany(b, coll)
}

func BenchmarkCollection_DeleteOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionDeleteOne(b, coll)
}

func BenchmarkCollection_DeleteMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollectionDeleteMany(b, coll)
}

func BenchmarkCollection_FindOne(b *testing.B) {
	b.Run("with index", func(b *testing.B) {
		server := Server()
		defer ReleaseServer(server)

		coll, err := testCollection(server)
		assert.NoError(b, err)

		databasetest.BenchmarkCollectionFindOneWithIndex(b, coll)
	})

	b.Run("without index", func(b *testing.B) {
		server := Server()
		defer ReleaseServer(server)

		coll, err := testCollection(server)
		assert.NoError(b, err)

		databasetest.BenchmarkCollectionFindOneWithoutIndex(b, coll)
	})
}

func BenchmarkCollection_FindMany(b *testing.B) {
	b.Run("with index", func(b *testing.B) {
		server := Server()
		defer ReleaseServer(server)

		coll, err := testCollection(server)
		assert.NoError(b, err)

		databasetest.BenchmarkCollectionFindManyWithIndex(b, coll)
	})

	b.Run("without index", func(b *testing.B) {
		server := Server()
		defer ReleaseServer(server)

		coll, err := testCollection(server)
		assert.NoError(b, err)

		databasetest.BenchmarkCollectionFindManyWithoutIndex(b, coll)
	})
}

func testCollection(server *memongo.Server) (*Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(server.URI()))
	if err != nil {
		return nil, err
	}
	db := client.Database(faker.UUIDHyphenated())

	return NewCollection(db.Collection(faker.UUIDHyphenated())), nil
}

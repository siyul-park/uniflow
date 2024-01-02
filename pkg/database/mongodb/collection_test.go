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

	databasetest.TestCollection_Name(t, coll)
}

func TestCollection_Indexes(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_Indexes(t, coll)
}

func TestCollection_Watch(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_Watch(t, coll)
}

func TestCollection_InsertOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_Insert(t, coll)
}

func TestCollection_InsertMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_InsertMany(t, coll)
}

func TestCollection_UpdateOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_UpdateOne(t, coll)
}

func TestCollection_UpdateMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_UpdateMany(t, coll)
}

func TestCollection_DeleteOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_DeleteOne(t, coll)
}

func TestCollection_DeleteMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_DeleteMany(t, coll)
}

func TestCollection_FindOne(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_FindOne(t, coll)
}

func TestCollection_FindMany(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_FindMany(t, coll)
}

func TestCollection_Drop(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(t, err)

	databasetest.TestCollection_Drop(t, coll)
}

func BenchmarkCollection_InsertOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_InsertOne(b, coll)
}

func BenchmarkCollection_InsertMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_InsertMany(b, coll)
}

func BenchmarkCollection_UpdateOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_UpdateOne(b, coll)
}

func BenchmarkCollection_UpdateMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_UpdateMany(b, coll)
}

func BenchmarkCollection_DeleteOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_DeleteOne(b, coll)
}

func BenchmarkCollection_DeleteMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_DeleteMany(b, coll)
}

func BenchmarkCollection_FindOne(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_FindOne(b, coll)
}

func BenchmarkCollection_FindMany(b *testing.B) {
	server := Server()
	defer ReleaseServer(server)

	coll, err := testCollection(server)
	assert.NoError(b, err)

	databasetest.BenchmarkCollection_FindMany(b, coll)
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

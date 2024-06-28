package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database/databasetest"
	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestDatabase_Name(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	db, err := testDatabase(server)
	assert.NoError(t, err)

	databasetest.TestDatabase_Name(t, db)
}

func TestDatabase_Collection(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	db, err := testDatabase(server)
	assert.NoError(t, err)

	databasetest.TestDatabase_Collection(t, db)
}

func TestDatabase_Drop(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	db, err := testDatabase(server)
	assert.NoError(t, err)

	databasetest.TestDatabase_Drop(t, db)
}

func testDatabase(server *memongo.Server) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(server.URI()))
	if err != nil {
		return nil, err
	}

	return newDatabase(client.Database(faker.UUIDHyphenated())), nil
}

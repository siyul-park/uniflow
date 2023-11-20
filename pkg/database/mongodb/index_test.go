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

func TestIndexView_List(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	indexView, err := testIndexView(server)
	assert.NoError(t, err)

	databasetest.AssertIndexViewList(t, indexView)
}

func TestIndexView_Create(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	indexView, err := testIndexView(server)
	assert.NoError(t, err)

	databasetest.AssertIndexViewCreate(t, indexView)
}

func TestIndexView_Drop(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	indexView, err := testIndexView(server)
	assert.NoError(t, err)

	databasetest.AssertIndexViewDrop(t, indexView)
}

func testIndexView(server *memongo.Server) (*IndexView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(server.URI()))
	if err != nil {
		return nil, err
	}

	db := client.Database(faker.UUIDHyphenated())
	coll := db.Collection(faker.UUIDHyphenated())

	return UpgradeIndexView(coll.Indexes()), nil
}

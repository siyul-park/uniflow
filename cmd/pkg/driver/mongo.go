package driver

import (
	"context"
	"strings"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"

	"github.com/siyul-park/uniflow/pkg/store"

	"github.com/gofrs/uuid"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongostore "github.com/siyul-park/uniflow/driver/mongo/pkg/store"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoDriver represents a MongoDB connection and provides methods to interact with the database.
type MongoDriver struct {
	server   *memongo.Server
	client   *mongo.Client
	database *mongo.Database
}

var _ Driver = (*MongoDriver)(nil)

// NewMongoDriver initializes a new MongoDB connection and returns a Driver instance.
func NewMongoDriver(uri, name string) (Driver, error) {
	var server *memongo.Server
	if strings.HasPrefix(uri, "memongodb://") {
		server = mongoserver.New()
		uri = server.URI()
	}

	if name == "" {
		name = uuid.Must(uuid.NewV7()).String()
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	return &MongoDriver{
		server:   server,
		client:   client,
		database: client.Database(name),
	}, nil
}

// NewSpecStore creates and returns a new Spec Store.
func (d *MongoDriver) NewSpecStore(ctx context.Context, name string) (store.Store, error) {
	if name == "" {
		name = "specs"
	}
	s := mongostore.New(d.database.Collection(name))
	err := s.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": true}},
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// NewValueStore creates and returns a new Value Store.
func (d *MongoDriver) NewValueStore(ctx context.Context, name string) (store.Store, error) {
	if name == "" {
		name = "values"
	}
	s := mongostore.New(d.database.Collection(name))
	err := s.Index(ctx, []string{value.KeyNamespace, value.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": true}},
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes the MongoDB connection.
func (d *MongoDriver) Close(ctx context.Context) error {
	if d.server != nil {
		defer mongoserver.Release(d.server)
	}
	return d.client.Disconnect(ctx)
}

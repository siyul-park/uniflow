package driver

import (
	"context"
	"strings"

	"github.com/gofrs/uuid"
	mongochart "github.com/siyul-park/uniflow/driver/mongo/pkg/chart"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongospec "github.com/siyul-park/uniflow/driver/mongo/pkg/spec"
	mongovalue "github.com/siyul-park/uniflow/driver/mongo/pkg/value"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
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

// ChartStore creates and returns a new Chart Store.
func (d *MongoDriver) ChartStore(ctx context.Context, name string) (chart.Store, error) {
	if name == "" {
		name = "charts"
	}

	collection := d.database.Collection(name)
	store := mongochart.NewStore(collection)

	if err := store.Index(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// SpecStore creates and returns a new Spec Store.
func (d *MongoDriver) SpecStore(ctx context.Context, name string) (spec.Store, error) {
	if name == "" {
		name = "specs"
	}

	collection := d.database.Collection(name)
	store := mongospec.NewStore(collection)

	if err := store.Index(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// ValueStore creates and returns a new Value Store.
func (d *MongoDriver) ValueStore(ctx context.Context, name string) (value.Store, error) {
	if name == "" {
		name = "values"
	}

	collection := d.database.Collection(name)
	store := mongovalue.NewStore(collection)

	if err := store.Index(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// Close closes the MongoDB connection.
func (d *MongoDriver) Close(ctx context.Context) error {
	if d.server != nil {
		defer mongoserver.Release(d.server)
	}
	return d.client.Disconnect(ctx)
}

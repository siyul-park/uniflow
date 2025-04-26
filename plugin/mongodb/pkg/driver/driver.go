package driver

import (
	"github.com/siyul-park/uniflow/pkg/driver"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

// Driver implements the driver.Driver interface for MongoDB.
type Driver struct{}

var _ driver.Driver = (*Driver)(nil)

func New() *Driver {
	return &Driver{}
}

// Open establishes a connection to MongoDB using the provided connection string.
func (d *Driver) Open(name string) (driver.Conn, error) {
	dsn, err := connstring.Parse(name)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(
		options.Client().
			ApplyURI(name).
			SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)),
	)
	if err != nil {
		return nil, err
	}

	return newConn(client, dsn.Database), nil
}

// Close performs any necessary cleanup.
func (d *Driver) Close() error {
	return nil
}

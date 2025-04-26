package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/driver"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type conn struct {
	client   *mongo.Client
	database *mongo.Database
}

var _ driver.Conn = (*conn)(nil)

func newConn(client *mongo.Client, database string) *conn {
	return &conn{client: client, database: client.Database(database)}
}

func (c *conn) Load(name string) (driver.Store, error) {
	return NewStore(c.database.Collection(name)), nil
}

func (c *conn) Close() error {
	return c.client.Disconnect(context.Background())
}

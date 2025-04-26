package driver

import (
	"github.com/siyul-park/uniflow/pkg/driver"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Conn implements the driver.Conn interface for a MongoDB connection.
type Conn struct {
	database *mongo.Database
}

var _ driver.Conn = (*Conn)(nil)

// Load returns a new Store for the given collection name.
func (c *Conn) Load(name string) (driver.Store, error) {
	return NewStore(c.database.Collection(name)), nil
}

// Close performs cleanup of the connection.
func (c *Conn) Close() error {
	return nil
}

package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	raw       *mongo.Client
	databases map[string]*Database
	lock      sync.RWMutex
}

func Connect(ctx context.Context, opts ...*options.ClientOptions) (*Connection, error) {
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return NewConnection(client), nil
}

func NewConnection(client *mongo.Client) *Connection {
	return &Connection{
		raw:       client,
		databases: map[string]*Database{},
	}
}

func (c *Connection) Database(_ context.Context, name string) (database.Database, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if db, ok := c.databases[name]; ok {
		return db, nil
	}

	db := newDatabase(c.raw.Database(name))
	c.databases[name] = db

	return db, nil
}

func (c *Connection) Disconnect(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.databases = map[string]*Database{}

	return c.raw.Disconnect(ctx)
}

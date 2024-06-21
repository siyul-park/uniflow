package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	internal  *mongo.Client
	databases map[string]*Database
	mu        sync.RWMutex
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
		internal:  client,
		databases: map[string]*Database{},
	}
}

func (c *Connection) Database(_ context.Context, name string) (database.Database, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if db, ok := c.databases[name]; ok {
		return db, nil
	}

	db := newDatabase(c.internal.Database(name))
	c.databases[name] = db

	return db, nil
}

func (c *Connection) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.databases = map[string]*Database{}

	return c.internal.Disconnect(ctx)
}

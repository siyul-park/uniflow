package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connection represents a MongoDB client connection manager.
type Connection struct {
	internal  *mongo.Client        
	databases map[string]*Database 
	mu        sync.RWMutex         
}

// Connect creates a new MongoDB connection using the provided options.
func Connect(ctx context.Context, opts ...*options.ClientOptions) (*Connection, error) {
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return NewConnection(client), nil
}

// NewConnection creates a new Connection instance with the given MongoDB client.
func NewConnection(client *mongo.Client) *Connection {
	return &Connection{
		internal:  client,
		databases: make(map[string]*Database),
	}
}

// Database returns a database handle for the specified database name.
func (c *Connection) Database(ctx context.Context, name string) (database.Database, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if db, ok := c.databases[name]; ok {
		return db, nil
	}

	db := newDatabase(c.internal.Database(name))
	c.databases[name] = db

	return db, nil
}

// Disconnect closes the connection to the MongoDB server and clears cached databases.
func (c *Connection) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.databases = make(map[string]*Database)

	return c.internal.Disconnect(ctx)
}

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

func (con *Connection) Database(_ context.Context, name string) (database.Database, error) {
	con.lock.Lock()
	defer con.lock.Unlock()

	if db, ok := con.databases[name]; ok {
		return db, nil
	}

	db := newDatabase(con.raw.Database(name))
	con.databases[name] = db

	return db, nil
}

func (con *Connection) Disconnect(ctx context.Context) error {
	con.lock.Lock()
	defer con.lock.Unlock()

	con.databases = map[string]*Database{}

	return con.raw.Disconnect(ctx)
}

package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	raw         *mongo.Database
	collections map[string]*Collection
	lock        sync.RWMutex
}

var _ database.Database = &Database{}

func NewDatabase(db *mongo.Database) *Database {
	return &Database{
		raw:         db,
		collections: map[string]*Collection{},
	}
}

func (db *Database) Name() string {
	return db.raw.Name()
}

func (db *Database) Collection(_ context.Context, name string) (database.Collection, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if coll, ok := db.collections[name]; ok {
		return coll, nil
	}

	coll := NewCollection(db.raw.Collection(name))
	db.collections[name] = coll

	return coll, nil
}

func (db *Database) Drop(ctx context.Context) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	for _, coll := range db.collections {
		if err := coll.Drop(ctx); err != nil {
			return err
		}
	}

	db.collections = map[string]*Collection{}

	return db.raw.Drop(ctx)
}

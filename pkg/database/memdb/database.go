package memdb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
)

type Database struct {
	name        string
	collections map[string]*Collection
	lock        sync.RWMutex
}

var _ database.Database = &Database{}

func New(name string) *Database {
	return &Database{
		name:        name,
		collections: map[string]*Collection{},
		lock:        sync.RWMutex{},
	}
}

func (db *Database) Name() string {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.name
}

func (db *Database) Collection(_ context.Context, name string) (database.Collection, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if coll, ok := db.collections[name]; ok {
		return coll, nil
	}

	coll := NewCollection(name)
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

	return nil
}

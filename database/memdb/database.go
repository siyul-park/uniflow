package memdb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/database"
)

// Database represents a database containing collections of documents.
type Database struct {
	name        string
	collections map[string]*Collection
	lock        sync.RWMutex
}

var _ database.Database = (*Database)(nil)

// New creates a new Database instance with the given name.
func New(name string) *Database {
	return &Database{
		name:        name,
		collections: map[string]*Collection{},
		lock:        sync.RWMutex{},
	}
}

// Name returns the name of the database.
func (d *Database) Name() string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.name
}

// Collection returns a Collection instance from the database with the given name.
// If the collection does not exist, it creates a new one and adds it to the database.
func (d *Database) Collection(_ context.Context, name string) (database.Collection, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if coll, ok := d.collections[name]; ok {
		return coll, nil
	}

	coll := NewCollection(name)
	d.collections[name] = coll

	return coll, nil
}

// Drop deletes all collections within the database and clears the collection map.
func (d *Database) Drop(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, coll := range d.collections {
		if err := coll.Drop(ctx); err != nil {
			return err
		}
	}

	d.collections = map[string]*Collection{}

	return nil
}

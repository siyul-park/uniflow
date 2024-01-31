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

var _ database.Database = (*Database)(nil)

func New(name string) *Database {
	return &Database{
		name:        name,
		collections: map[string]*Collection{},
		lock:        sync.RWMutex{},
	}
}

func (d *Database) Name() string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.name
}

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

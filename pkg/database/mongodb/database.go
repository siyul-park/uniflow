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

func newDatabase(db *mongo.Database) *Database {
	return &Database{
		raw:         db,
		collections: map[string]*Collection{},
	}
}

func (d *Database) Name() string {
	return d.raw.Name()
}

func (d *Database) Collection(_ context.Context, name string) (database.Collection, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if coll, ok := d.collections[name]; ok {
		return coll, nil
	}

	coll := newCollection(d.raw.Collection(name))
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

	return d.raw.Drop(ctx)
}

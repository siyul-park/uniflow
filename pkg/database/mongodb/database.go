package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
)

// Database represents a MongoDB database manager.
type Database struct {
	internal    *mongo.Database
	collections map[string]*Collection
	mu          sync.RWMutex
}

var _ database.Database = (*Database)(nil)

func newDatabase(db *mongo.Database) *Database {
	return &Database{
		internal:    db,
		collections: make(map[string]*Collection),
	}
}

// Name returns the name of the MongoDB database.
func (d *Database) Name() string {
	return d.internal.Name()
}

// Collection returns a collection handle for the specified collection name.
func (d *Database) Collection(ctx context.Context, name string) (database.Collection, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if coll, ok := d.collections[name]; ok {
		return coll, nil
	}

	coll := newCollection(d.internal.Collection(name))
	d.collections[name] = coll

	return coll, nil
}

// Drop deletes the MongoDB database and clears cached collections.
func (d *Database) Drop(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Drop all collections in the database first.
	for _, coll := range d.collections {
		if err := coll.Drop(ctx); err != nil {
			return err
		}
	}

	// Clear the collections map to release references.
	d.collections = make(map[string]*Collection)

	// Drop the actual MongoDB database.
	return d.internal.Drop(ctx)
}

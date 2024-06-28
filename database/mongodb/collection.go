package mongodb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/object"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection represents a MongoDB collection wrapper that implements the database.Collection interface.
type Collection struct {
	internal *mongo.Collection
	mu       sync.RWMutex
}

var _ database.Collection = (*Collection)(nil)

func newCollection(coll *mongo.Collection) *Collection {
	return &Collection{internal: coll}
}

// Name returns the name of the collection.
func (c *Collection) Name() string {
	return c.internal.Name()
}

// Indexes returns an IndexView for managing indexes on this collection.
func (c *Collection) Indexes() database.IndexView {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return newIndexView(c.internal.Indexes())
}

// Watch starts a change stream on the collection based on the provided filter.
func (c *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var pipeline mongo.Pipeline
	if filter != nil {
		if match, err := filterToBson(filter); err != nil {
			return nil, err
		} else if match != nil {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
		}
	}

	stream, err := c.internal.Watch(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	return newStream(ctx, stream), nil
}

// InsertOne inserts a single document into the collection.
func (c *Collection) InsertOne(ctx context.Context, doc object.Map) (object.Object, error) {
	raw, err := primitiveToBson(doc)
	if err != nil {
		return nil, err
	}

	res, err := c.internal.InsertOne(ctx, raw)
	if err != nil {
		return nil, errors.Wrap(database.ErrWrite, err.Error())
	}

	var id object.Object
	if err := bsonToPrimitive(res.InsertedID, &id); err != nil {
		return nil, err
	}
	return id, nil
}

// InsertMany inserts multiple documents into the collection.
func (c *Collection) InsertMany(ctx context.Context, docs []object.Map) ([]object.Object, error) {
	var raws bson.A
	for _, doc := range docs {
		if raw, err := primitiveToBson(doc); err != nil {
			return nil, err
		} else {
			raws = append(raws, raw)
		}
	}

	res, err := c.internal.InsertMany(ctx, raws)
	if err != nil {
		return nil, errors.Wrap(database.ErrWrite, err.Error())
	}

	var ids []object.Object
	for _, insertedID := range res.InsertedIDs {
		var id object.Object
		if err := bsonToPrimitive(insertedID, &id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// UpdateOne updates a single document in the collection matching the filter with the provided patch.
func (c *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch object.Map, opts ...*database.UpdateOptions) (bool, error) {
	raw, err := primitiveToBson(patch)
	if err != nil {
		return false, err
	}
	f, err := filterToBson(filter)
	if err != nil {
		return false, err
	}

	res, err := c.internal.UpdateOne(ctx, f, bson.M{"$set": raw}, internalUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return false, errors.Wrap(database.ErrWrite, err.Error())
	}

	return res.UpsertedCount+res.ModifiedCount > 0, nil
}

// UpdateMany updates multiple documents in the collection matching the filter with the provided patch.
func (c *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch object.Map, opts ...*database.UpdateOptions) (int, error) {
	raw, err := primitiveToBson(patch)
	if err != nil {
		return 0, err
	}
	f, err := filterToBson(filter)
	if err != nil {
		return 0, err
	}

	res, err := c.internal.UpdateMany(ctx, f, bson.M{"$set": raw}, internalUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return 0, errors.Wrap(database.ErrWrite, err.Error())
	}

	return int(res.UpsertedCount + res.ModifiedCount), nil
}

// DeleteOne deletes a single document from the collection matching the filter.
func (c *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	f, err := filterToBson(filter)
	if err != nil {
		return false, err
	}

	res, err := c.internal.DeleteOne(ctx, f)
	if err != nil {
		return false, errors.Wrap(database.ErrDelete, err.Error())
	}

	return res.DeletedCount > 0, nil
}

// DeleteMany deletes multiple documents from the collection matching the filter.
func (c *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	f, err := filterToBson(filter)
	if err != nil {
		return 0, err
	}

	res, err := c.internal.DeleteMany(ctx, f)
	if err != nil {
		return 0, errors.Wrap(database.ErrDelete, err.Error())
	}

	return int(res.DeletedCount), nil
}

// FindOne retrieves a single document from the collection matching the filter.
func (c *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (object.Map, error) {
	f, err := filterToBson(filter)
	if err != nil {
		return nil, err
	}

	res := c.internal.FindOne(ctx, f, internalFindOneOptions(database.MergeFindOptions(opts)))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.Wrap(database.ErrRead, res.Err().Error())
	}

	var doc object.Object
	var r any
	if err := res.Decode(&r); err != nil {
		return nil, err
	}
	if err := bsonToPrimitive(r, &doc); err != nil {
		return nil, err
	}
	return doc.(object.Map), nil
}

// FindMany retrieves multiple documents from the collection matching the filter.
func (c *Collection) FindMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]object.Map, error) {
	f, err := filterToBson(filter)
	if err != nil {
		return nil, err
	}

	cursor, err := c.internal.Find(ctx, f, internalFindOptions(database.MergeFindOptions(opts)))
	if err != nil {
		return nil, errors.Wrap(database.ErrRead, err.Error())
	}

	var docs []object.Map
	for cursor.Next(ctx) {
		var doc object.Object
		var r any
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		if err := bsonToPrimitive(r, &doc); err != nil {
			return nil, err
		}
		docs = append(docs, doc.(object.Map))
	}

	return docs, nil
}

// Drop drops the entire collection.
func (c *Collection) Drop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.internal.Drop(ctx); err != nil {
		return errors.Wrap(database.ErrDelete, err.Error())
	}

	return nil
}

func internalUpdateOptions(opts *database.UpdateOptions) *options.UpdateOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.UpdateOptions{
		Upsert: opts.Upsert,
	})
}

func internalFindOneOptions(opts *database.FindOptions) *options.FindOneOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOneOptions{
		Skip: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort: sortToBson(opts.Sorts),
	})
}

func internalFindOptions(opts *database.FindOptions) *options.FindOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOptions{
		Limit: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Limit))),
		Skip:  lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort:  sortToBson(opts.Sorts),
	})
}

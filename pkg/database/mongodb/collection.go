package mongodb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	raw  *mongo.Collection
	lock sync.RWMutex
}

var _ database.Collection = &Collection{}

func newCollection(coll *mongo.Collection) *Collection {
	return &Collection{raw: coll}
}

func (c *Collection) Name() string {
	return c.raw.Name()
}

func (c *Collection) Indexes() database.IndexView {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return newIndexView(c.raw.Indexes())
}

func (c *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	pipeline := mongo.Pipeline{}

	if filter != nil {
		if match, err := marshalFilter(filter); err != nil {
			return nil, err
		} else if match != nil {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
		}
	}

	stream, err := c.raw.Watch(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	return newStream(ctx, stream), nil
}

func (c *Collection) InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error) {
	raw, err := marshalDocument(doc)
	if err != nil {
		return nil, err
	}

	res, err := c.raw.InsertOne(ctx, raw)
	if err != nil {
		return nil, errors.Wrap(database.ErrWrite, err.Error())
	}

	var id primitive.Value
	if err := unmarshalDocument(res.InsertedID, &id); err != nil {
		return nil, err
	}
	return id, nil
}

func (c *Collection) InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	var raws bson.A
	for _, doc := range docs {
		if raw, err := marshalDocument(doc); err == nil {
			raws = append(raws, raw)
		} else {
			return nil, err
		}
	}

	res, err := c.raw.InsertMany(ctx, raws)
	if err != nil {
		return nil, errors.Wrap(database.ErrWrite, err.Error())
	}

	var ids []primitive.Value
	for _, insertedID := range res.InsertedIDs {
		var id primitive.Value
		if err := unmarshalDocument(insertedID, &id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (c *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (bool, error) {
	raw, err := marshalDocument(patch)
	if err != nil {
		return false, err
	}
	f, err := marshalFilter(filter)
	if err != nil {
		return false, err
	}

	res, err := c.raw.UpdateOne(ctx, f, bson.M{"$set": raw}, marshalUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return false, errors.Wrap(database.ErrWrite, err.Error())
	}

	return res.UpsertedCount+res.ModifiedCount > 0, nil
}

func (c *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (int, error) {
	raw, err := marshalDocument(patch)
	if err != nil {
		return 0, err
	}
	f, err := marshalFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := c.raw.UpdateMany(ctx, f, bson.M{"$set": raw}, marshalUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return 0, errors.Wrap(database.ErrWrite, err.Error())
	}

	return int(res.UpsertedCount + res.ModifiedCount), nil
}

func (c *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return false, err
	}

	res, err := c.raw.DeleteOne(ctx, f)
	if err != nil {
		return false, errors.Wrap(database.ErrDelete, err.Error())
	}

	return res.DeletedCount > 0, nil
}

func (c *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := c.raw.DeleteMany(ctx, f)
	if err != nil {
		return 0, errors.Wrap(database.ErrDelete, err.Error())
	}

	return int(res.DeletedCount), nil
}

func (c *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return nil, err
	}

	res := c.raw.FindOne(ctx, f, marshalFindOneOptions(database.MergeFindOptions(opts)))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.Wrap(database.ErrRead, res.Err().Error())
	}

	var doc primitive.Value
	var r any
	if err := res.Decode(&r); err != nil {
		return nil, err
	}
	if err := unmarshalDocument(r, &doc); err != nil {
		return nil, err
	}
	return doc.(*primitive.Map), nil
}

func (c *Collection) FindMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return nil, err
	}

	cursor, err := c.raw.Find(ctx, f, marshalFindOptions(database.MergeFindOptions(opts)))
	if err != nil {
		return nil, errors.Wrap(database.ErrRead, err.Error())
	}

	var docs []*primitive.Map
	for cursor.Next(ctx) {
		var doc primitive.Value
		var r any
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		if err := unmarshalDocument(r, &doc); err != nil {
			return nil, err
		}
		docs = append(docs, doc.(*primitive.Map))
	}

	return docs, nil
}

func (c *Collection) Drop(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.raw.Drop(ctx); err != nil {
		return errors.Wrap(database.ErrDelete, err.Error())
	}

	return nil
}

func marshalUpdateOptions(opts *database.UpdateOptions) *options.UpdateOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.UpdateOptions{
		Upsert: opts.Upsert,
	})
}

func marshalFindOneOptions(opts *database.FindOptions) *options.FindOneOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOneOptions{
		Skip: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort: marshalSorts(opts.Sorts),
	})
}

func marshalFindOptions(opts *database.FindOptions) *options.FindOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOptions{
		Limit: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Limit))),
		Skip:  lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort:  marshalSorts(opts.Sorts),
	})
}

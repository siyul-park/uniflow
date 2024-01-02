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

func NewCollection(coll *mongo.Collection) *Collection {
	return &Collection{raw: coll}
}

func (coll *Collection) Name() string {
	return coll.raw.Name()
}

func (coll *Collection) Indexes() database.IndexView {
	coll.lock.RLock()
	defer coll.lock.RUnlock()

	return UpgradeIndexView(coll.raw.Indexes())
}

func (coll *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	coll.lock.Lock()
	defer coll.lock.Unlock()

	pipeline := mongo.Pipeline{}

	if filter != nil {
		if match, err := marshalFilter(filter); err != nil {
			return nil, err
		} else if match != nil {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
		}
	}

	stream, err := coll.raw.Watch(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	return UpgradeStream(ctx, stream), nil
}

func (coll *Collection) InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error) {
	raw, err := marshalDocument(doc)
	if err != nil {
		return nil, err
	}

	res, err := coll.raw.InsertOne(ctx, raw)
	if err != nil {
		return nil, errors.Wrap(database.ErrWrite, err.Error())
	}

	var id primitive.Value
	if err := unmarshalDocument(res.InsertedID, &id); err != nil {
		return nil, err
	}
	return id, nil
}

func (coll *Collection) InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	var raws bson.A
	for _, doc := range docs {
		if raw, err := marshalDocument(doc); err == nil {
			raws = append(raws, raw)
		} else {
			return nil, err
		}
	}

	res, err := coll.raw.InsertMany(ctx, raws)
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

func (coll *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (bool, error) {
	raw, err := marshalDocument(patch)
	if err != nil {
		return false, err
	}
	f, err := marshalFilter(filter)
	if err != nil {
		return false, err
	}

	res, err := coll.raw.UpdateOne(ctx, f, bson.M{"$set": raw}, mongoUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return false, errors.Wrap(database.ErrWrite, err.Error())
	}

	return res.UpsertedCount+res.ModifiedCount > 0, nil
}

func (coll *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (int, error) {
	raw, err := marshalDocument(patch)
	if err != nil {
		return 0, err
	}
	f, err := marshalFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := coll.raw.UpdateMany(ctx, f, bson.M{"$set": raw}, mongoUpdateOptions(database.MergeUpdateOptions(opts)))
	if err != nil {
		return 0, errors.Wrap(database.ErrWrite, err.Error())
	}

	return int(res.UpsertedCount + res.ModifiedCount), nil
}

func (coll *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return false, err
	}

	res, err := coll.raw.DeleteOne(ctx, f)
	if err != nil {
		return false, errors.Wrap(database.ErrDelete, err.Error())
	}

	return res.DeletedCount > 0, nil
}

func (coll *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := coll.raw.DeleteMany(ctx, f)
	if err != nil {
		return 0, errors.Wrap(database.ErrDelete, err.Error())
	}

	return int(res.DeletedCount), nil
}

func (coll *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return nil, err
	}

	res := coll.raw.FindOne(ctx, f, mongoFindOneOptions(database.MergeFindOptions(opts)))
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

func (coll *Collection) FindMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	f, err := marshalFilter(filter)
	if err != nil {
		return nil, err
	}

	cursor, err := coll.raw.Find(ctx, f, mongoFindOptions(database.MergeFindOptions(opts)))
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

func (coll *Collection) Drop(ctx context.Context) error {
	coll.lock.Lock()
	defer coll.lock.Unlock()

	if err := coll.raw.Drop(ctx); err != nil {
		return errors.Wrap(database.ErrDelete, err.Error())
	}

	return nil
}

func mongoUpdateOptions(opts *database.UpdateOptions) *options.UpdateOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.UpdateOptions{
		Upsert: opts.Upsert,
	})
}

func mongoFindOneOptions(opts *database.FindOptions) *options.FindOneOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOneOptions{
		Skip: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort: mongoSorts(opts.Sorts),
	})
}

func mongoFindOptions(opts *database.FindOptions) *options.FindOptions {
	if opts == nil {
		return nil
	}
	return lo.ToPtr(options.FindOptions{
		Limit: lo.EmptyableToPtr(int64(lo.FromPtr(opts.Limit))),
		Skip:  lo.EmptyableToPtr(int64(lo.FromPtr(opts.Skip))),
		Sort:  mongoSorts(opts.Sorts),
	})
}

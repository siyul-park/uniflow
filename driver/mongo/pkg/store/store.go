package store

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	collection *mongo.Collection
}

var _ store.Store = (*Store)(nil)

func New(collection *mongo.Collection) *Store {
	return &Store{collection: collection}
}

func (s *Store) Watch(ctx context.Context, filter any) (store.Stream, error) {
	f, err := types.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filter, err = toBSON(f)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{}
	if filter != nil {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})
	}

	cs, err := s.collection.Watch(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	return &stream{changeStream: cs}, nil
}

func (s *Store) Index(ctx context.Context, keys []string, opts ...store.IndexOptions) error {
	option := options.Index()
	for _, opt := range opts {
		if opt.Unique {
			option = option.SetUnique(opt.Unique)
		}
		if opt.Filter != nil {
			val, err := types.Marshal(opt.Filter)
			if err != nil {
				return err
			}

			raw, err := toBSON(val)
			if err != nil {
				return err
			}

			option = option.SetPartialFilterExpression(raw)
		}
	}

	name := ""
	for i, key := range keys {
		if key == "id" {
			key = "_id"
		}
		if i > 0 {
			name += "_"
		}
		name += key + "_1"
	}

	idx := bson.D{}
	for _, key := range keys {
		if key == "id" {
			key = "_id"
		}
		idx = append(idx, bson.E{Key: key, Value: 1})
	}

	indexes, err := s.collection.Indexes().List(ctx)
	if err != nil {
		return err
	}

	for indexes.Next(ctx) {
		var index mongo.IndexSpecification
		if err := indexes.Decode(&index); err != nil {
			return err
		}

		if index.Name == name {
			return nil
		}
	}

	_, err = s.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    idx,
		Options: option,
	})
	return err
}

func (s *Store) Unindex(ctx context.Context, keys []string) error {
	name := ""
	for i, key := range keys {
		if key == "id" {
			key = "_id"
		}
		if i > 0 {
			name += "_"
		}
		name += key + "_1"
	}

	indexes, err := s.collection.Indexes().List(ctx)
	if err != nil {
		return err
	}

	for indexes.Next(ctx) {
		var index mongo.IndexSpecification
		if err := indexes.Decode(&index); err != nil {
			return err
		}

		if index.Name == name {
			return s.collection.Indexes().DropOne(ctx, name)
		}
	}

	return nil
}

func (s *Store) Insert(ctx context.Context, docs []any, _ ...store.InsertOptions) error {
	raws := make([]any, 0, len(docs))
	for _, doc := range docs {
		val, err := types.Marshal(doc)
		if err != nil {
			return err
		}

		raw, err := toBSON(val)
		if err != nil {
			return err
		}
		raws = append(raws, raw)
	}

	_, err := s.collection.InsertMany(ctx, raws)
	return err
}

func (s *Store) Update(ctx context.Context, filter, update any, opts ...store.UpdateOptions) (int, error) {
	option := options.UpdateMany()
	for _, opt := range opts {
		if opt.Upsert {
			option = option.SetUpsert(opt.Upsert)
		}
	}

	f, err := types.Marshal(filter)
	if err != nil {
		return 0, err
	}
	filter, err = toBSON(f)
	if err != nil {
		return 0, err
	}

	u, err := types.Marshal(update)
	if err != nil {
		return 0, err
	}
	update, err = toBSON(u)
	if err != nil {
		return 0, err
	}

	res, err := s.collection.UpdateMany(ctx, filter, update, option)
	if err != nil {
		return 0, err
	}
	return int(res.ModifiedCount + res.UpsertedCount), nil
}

func (s *Store) Delete(ctx context.Context, filter any, _ ...store.DeleteOptions) (int, error) {
	f, err := types.Marshal(filter)
	if err != nil {
		return 0, err
	}
	filter, err = toBSON(f)
	if err != nil {
		return 0, err
	}

	res, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(res.DeletedCount), nil
}

func (s *Store) Find(ctx context.Context, filter any, opts ...store.FindOptions) (store.Cursor, error) {
	option := options.Find()
	for _, opt := range opts {
		if opt.Limit != 0 {
			option = option.SetLimit(int64(opt.Limit))
		}
		if opt.Sort != nil {
			val, err := types.Marshal(opt.Sort)
			if err != nil {
				return nil, err
			}
			sort, err := toBSON(val)
			if err != nil {
				return nil, err
			}
			option = option.SetSort(sort)
		}
	}

	if filter == nil {
		filter = map[string]any{}
	}

	f, err := types.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filter, err = toBSON(f)
	if err != nil {
		return nil, err
	}

	cur, err := s.collection.Find(ctx, filter, option)
	if err != nil {
		return nil, err
	}
	return &cursor{cursor: cur}, nil
}

package mongodb

import (
	"context"

	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	IndexView struct {
		raw mongo.IndexView
	}
)

var _ database.IndexView = &IndexView{}

func UpgradeIndexView(iv mongo.IndexView) *IndexView {
	return &IndexView{raw: iv}
}

func (iv *IndexView) List(ctx context.Context) ([]database.IndexModel, error) {
	cursor, err := iv.raw.List(ctx)
	if err != nil {
		return nil, err
	}

	var indexes []bson.M
	if err := cursor.All(context.Background(), &indexes); err != nil {
		return nil, err
	}

	var models []database.IndexModel
	for _, index := range indexes {
		key, _ := index["key"].(bson.M)
		name, _ := index["name"].(string)
		unique, _ := index["unique"].(bool)
		partialFilterExpression, _ := index["partialFilterExpression"].(bson.M)

		var keys []string
		for k := range key {
			keys = append(keys, documentKey(k))
		}
		var partial *database.Filter
		if err := UnmarshalFilter(partialFilterExpression, &partial); err != nil {
			return nil, err
		}

		models = append(models, database.IndexModel{
			Keys:    keys,
			Name:    name,
			Unique:  unique,
			Partial: partial,
		})
	}

	return models, nil
}

func (iv *IndexView) Create(ctx context.Context, index database.IndexModel) error {
	keys := bson.D{}
	for _, k := range index.Keys {
		keys = append(keys, bson.E{Key: bsonKey(k), Value: 1})
	}

	partialFilterExpression, err := MarshalFilter(index.Partial)
	if err != nil {
		return err
	}

	_, err = iv.raw.CreateOne(ctx, mongo.IndexModel{
		Keys: keys,
		Options: &options.IndexOptions{
			Name:                    util.Ptr(index.Name),
			Unique:                  util.Ptr(index.Unique),
			PartialFilterExpression: partialFilterExpression,
		},
	})

	return err
}

func (iv *IndexView) Drop(ctx context.Context, name string) error {
	_, err := iv.raw.DropOne(ctx, name)
	return err
}

package mongodb

import (
	"context"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexView struct {
	internal mongo.IndexView
}

var _ database.IndexView = &IndexView{}

func newIndexView(v mongo.IndexView) *IndexView {
	return &IndexView{internal: v}
}

func (v *IndexView) List(ctx context.Context) ([]database.IndexModel, error) {
	cursor, err := v.internal.List(ctx)
	if err != nil {
		return nil, err
	}

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
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
			keys = append(keys, externalKey(k))
		}
		var partial *database.Filter
		if err := bsonToFilter(partialFilterExpression, &partial); err != nil {
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

func (v *IndexView) Create(ctx context.Context, index database.IndexModel) error {
	keys := bson.D{}
	for _, k := range index.Keys {
		keys = append(keys, bson.E{Key: internalKey(k), Value: 1})
	}

	partialFilterExpression, err := filterToBson(index.Partial)
	if err != nil {
		return err
	}

	_, err = v.internal.CreateOne(ctx, mongo.IndexModel{
		Keys: keys,
		Options: &options.IndexOptions{
			Name:                    lo.ToPtr(index.Name),
			Unique:                  lo.ToPtr(index.Unique),
			PartialFilterExpression: partialFilterExpression,
		},
	})

	return err
}

func (v *IndexView) Drop(ctx context.Context, name string) error {
	_, err := v.internal.DropOne(ctx, name)
	return err
}

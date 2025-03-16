package store

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type cursor struct {
	cursor *mongo.Cursor
}

var _ store.Cursor = (*cursor)(nil)

func (c *cursor) All(ctx context.Context, val any) error {
	defer c.cursor.Close(ctx)

	var elements []types.Value
	for c.cursor.Next(ctx) {
		var raw any
		if err := c.cursor.Decode(&raw); err != nil {
			return err
		}
		v, err := fromBSON(raw)
		if err != nil {
			return err
		}
		elements = append(elements, v)
	}
	return types.Unmarshal(types.NewSlice(elements...), val)
}

func (c *cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

func (c *cursor) Decode(val any) error {
	var raw any
	if err := c.cursor.Decode(&raw); err != nil {
		return err
	}
	v, err := fromBSON(raw)
	if err != nil {
		return err
	}
	return types.Unmarshal(v, val)
}

func (c *cursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

package driver

import (
	"context"
	"encoding/json"

	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/xwb1989/sqlparser"
)

// Table represents a schema.Table backed by a driver.Store.
type Table[T meta.Meta] struct {
	store driver.Store
}

var _ schema.Table = (*Table[meta.Meta])(nil)

// NewTable creates a new instance of Table with the provided driver.Store.
func NewTable[T meta.Meta](store driver.Store) *Table[T] {
	return &Table[T]{store: store}
}

// Scan retrieves rows from the store and returns them as a schema.Cursor.
func (t *Table[T]) Scan(ctx context.Context) (schema.Cursor, error) {
	cursor, err := t.store.Find(ctx, nil)
	if err != nil {
		return nil, err
	}

	var rows []schema.Row
	for cursor.Next(ctx) {
		var doc T
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		raw, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, err
		}

		var row schema.Row
		for k, v := range data {
			col := &sqlparser.ColName{Name: sqlparser.NewColIdent(k)}
			val, err := schema.Marshal(v)
			if err != nil {
				return nil, err
			}
			row.Columns = append(row.Columns, col)
			row.Values = append(row.Values, val)
		}
		rows = append(rows, row)
	}
	return schema.NewInMemoryCursor(rows), nil
}

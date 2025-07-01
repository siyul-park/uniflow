package driver

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/siyul-park/sqlbridge/engine"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/types"
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

// Indexes returns the list of indexes for the table.
func (t *Table[T]) Indexes(ctx context.Context) ([]schema.Index, error) {
	keys, err := t.store.Indexes(ctx)
	if err != nil {
		return nil, err
	}

	indexes := make([]schema.Index, 0, len(keys))
	for _, ks := range keys {
		cols := make([]*sqlparser.ColName, 0, len(ks))
		for _, k := range ks {
			cols = append(cols, &sqlparser.ColName{Name: sqlparser.NewColIdent(k)})
		}
		indexes = append(indexes, schema.Index{Name: strings.Join(ks, "_"), Columns: cols})
	}
	return indexes, nil
}

// Scan retrieves rows from the store and returns them as a schema.Cursor.
func (t *Table[T]) Scan(ctx context.Context, hint ...schema.ScanHint) (schema.Cursor, error) {
	keys, err := t.store.Indexes(ctx)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string][]string)
	for _, k := range keys {
		indexes[strings.Join(k, "_")] = k
	}

	filter := make(map[string]any)
	for _, h := range hint {
		ks, ok := indexes[h.Index]
		if !ok {
			continue
		}

		for i, r := range h.Ranges {
			if i >= len(ks) {
				break
			}
			field := ks[i]

			var from, to engine.Value
			if r.Min != nil {
				if v, err := engine.FromSQL(*r.Min); err != nil {
					return nil, err
				} else if from, err = t.valueOf(field, v.Interface()); err != nil {
					return nil, err
				}
			}
			if r.Max != nil {
				if v, err := engine.FromSQL(*r.Max); err != nil {
					return nil, err
				} else if to, err = t.valueOf(field, v.Interface()); err != nil {
					return nil, err
				}
			}

			cmp, err := engine.Compare(from, to)
			if err != nil {
				return nil, err
			}

			switch {
			case from != nil && to != nil && cmp == 0:
				filter[field] = from.Interface()
			case from != nil && to != nil:
				filter[field] = map[string]any{"$gte": from.Interface(), "$lte": to.Interface()}
			case from != nil:
				filter[field] = map[string]any{"$gte": from.Interface()}
			case to != nil:
				filter[field] = map[string]any{"$lte": to.Interface()}
			}
		}
	}

	cursor, err := t.store.Find(ctx, filter)
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

func (t *Table[T]) valueOf(key string, val any) (engine.Value, error) {
	data, err := types.Marshal(map[string]any{key: val})
	if err != nil {
		return nil, err
	}

	var doc T
	if err := types.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	data, err = types.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return engine.NewValue(types.InterfaceOf(types.Lookup(data, key))), nil
}

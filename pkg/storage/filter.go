package storage

import (
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Filter is a filter for finding matched document.
type Filter struct {
	OP       database.OP // Operator for the filter.
	Key      string      // Key specifies the field for the filter.
	Value    any         // Value is the filter value.
	Children []*Filter   // Children are nested filters for AND and OR operations.
}

type filterHelper[T any] struct {
	key string
}

// Where creates a new filterHelper with the specified key.
func Where[T any](key string) *filterHelper[T] {
	return &filterHelper[T]{key: key}
}

// EQ creates an equality filter.
func (fh *filterHelper[T]) EQ(value T) *Filter {
	return &Filter{
		OP:    database.EQ,
		Key:   fh.key,
		Value: value,
	}
}

// NE creates a not-equal filter.
func (fh *filterHelper[T]) NE(value T) *Filter {
	return &Filter{
		OP:    database.NE,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper[T]) LT(value T) *Filter {
	return &Filter{
		OP:    database.LT,
		Key:   fh.key,
		Value: value,
	}
}

// LTE creates a less-than-or-equal filter.
func (fh *filterHelper[T]) LTE(value T) *Filter {
	return &Filter{
		OP:    database.LTE,
		Key:   fh.key,
		Value: value,
	}
}

// GT creates a greater-than filter.
func (fh *filterHelper[T]) GT(value T) *Filter {
	return &Filter{
		OP:    database.GT,
		Key:   fh.key,
		Value: value,
	}
}

// GTE creates a greater-than-or-equal filter.
func (fh *filterHelper[T]) GTE(value T) *Filter {
	return &Filter{
		OP:    database.GTE,
		Key:   fh.key,
		Value: value,
	}
}

// IN creates a filter for values in a given slice.
func (fh *filterHelper[T]) IN(slice ...T) *Filter {
	value := make([]any, len(slice))
	for i, e := range slice {
		value[i] = e
	}
	return &Filter{
		OP:    database.IN,
		Key:   fh.key,
		Value: value,
	}
}

// NotIN creates a filter for values not in a given slice.
func (fh *filterHelper[T]) NotIN(slice ...T) *Filter {
	value := make([]any, len(slice))
	for i, e := range slice {
		value[i] = e
	}
	return &Filter{
		OP:    database.NIN,
		Key:   fh.key,
		Value: value,
	}
}

// IsNull creates a filter for null values.
func (fh *filterHelper[T]) IsNull() *Filter {
	return &Filter{
		OP:  database.NULL,
		Key: fh.key,
	}
}

// IsNotNull creates a filter for non-null values.
func (fh *filterHelper[T]) IsNotNull() *Filter {
	return &Filter{
		OP:  database.NNULL,
		Key: fh.key,
	}
}

// And creates a filter that combines multiple filters with a logical AND.
func (ft *Filter) And(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:       database.AND,
		Children: v,
	}
}

// Or creates a filter that combines multiple filters with a logical OR.
func (ft *Filter) Or(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:       database.OR,
		Children: v,
	}
}

// Encode encodes the filter to a database.Filter.
func (ft *Filter) Encode() (*database.Filter, error) {
	if ft == nil {
		return nil, nil
	}
	if ft.OP == database.AND || ft.OP == database.OR {
		var children []*database.Filter
		for _, child := range ft.Children {
			if c, err := child.Encode(); err != nil {
				return nil, err
			} else {
				children = append(children, c)
			}
		}
		return &database.Filter{OP: database.AND, Children: children}, nil
	}
	if ft.OP == database.NULL || ft.OP == database.NNULL {
		return &database.Filter{OP: ft.OP, Key: ft.Key}, nil
	}

	if v, err := primitive.MarshalBinary(ft.Value); err != nil {
		return nil, err
	} else {
		return &database.Filter{OP: ft.OP, Key: ft.Key, Value: v}, nil
	}
}

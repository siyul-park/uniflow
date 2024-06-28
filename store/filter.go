package store

import (
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/spec"
)

// Filter is a filter for finding matched document.
type Filter struct {
	OP       database.OP `map:"op"`                 // Operator for the filter.
	Key      string      `map:"key,omitempty"`      // Key specifies the field for the filter.
	Value    any         `map:"value,omitempty"`    // Value is the filter value.
	Children []*Filter   `map:"children,omitempty"` // Children are nested filters for AND OR operations.
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

// LT creates a less-than filter.
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

// IN creates an "in" filter.
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

// NotIN creates a "not in" filter.
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

// IsNull creates an "is null" filter.
func (fh *filterHelper[T]) IsNull() *Filter {
	return &Filter{
		OP:  database.NULL,
		Key: fh.key,
	}
}

// IsNotNull creates an "is not null" filter.
func (fh *filterHelper[T]) IsNotNull() *Filter {
	return &Filter{
		OP:  database.NNULL,
		Key: fh.key,
	}
}

// And creates an "and" filter.
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

// Or creates an "or" filter.
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
		return &database.Filter{OP: ft.OP, Children: children}, nil
	}

	if ft.OP == database.NULL || ft.OP == database.NNULL {
		return &database.Filter{OP: ft.OP, Key: ft.Key}, nil
	}

	value := ft.Value
	if ft.OP == database.IN || ft.OP == database.NIN {
		if v, err := object.MarshalBinary(ft.Value); err != nil {
			return nil, err
		} else if v, ok := v.(object.Slice); ok {
			elements := make([]any, 0, v.Len())
			for _, v := range v.Values() {
				unstructed := spec.NewUnstructured(nil)
				if err := unstructed.Set(ft.Key, v); err != nil {
					return nil, err
				} else if v, err := unstructed.Get(ft.Key); err != nil {
					return nil, err
				} else {
					elements = append(elements, v)
				}
			}
			value = elements
		}
	} else {
		unstructed := spec.NewUnstructured(nil)
		if err := unstructed.Set(ft.Key, ft.Value); err != nil {
			return nil, err
		} else if value, err = unstructed.Get(ft.Key); err != nil {
			return nil, err
		}
	}

	if v, err := object.MarshalBinary(value); err != nil {
		return nil, err
	} else {
		return &database.Filter{OP: ft.OP, Key: ft.Key, Value: v}, nil
	}
}

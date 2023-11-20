package storage

import (
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	// Filter is a filter for find matched primitive.
	Filter struct {
		OP    database.OP
		Key   string
		Value any
	}

	filterHelper[T any] struct {
		key string
	}
)

func Where[T any](key string) *filterHelper[T] {
	return &filterHelper[T]{
		key: key,
	}
}

func (fh *filterHelper[T]) EQ(value T) *Filter {
	return &Filter{
		OP:    database.EQ,
		Key:   fh.key,
		Value: value,
	}
}

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

func (fh *filterHelper[T]) LTE(value T) *Filter {
	return &Filter{
		OP:    database.LTE,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper[T]) GT(value T) *Filter {
	return &Filter{
		OP:    database.GT,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper[T]) GTE(value T) *Filter {
	return &Filter{
		OP:    database.GTE,
		Key:   fh.key,
		Value: value,
	}
}

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

func (fh *filterHelper[T]) IsNull() *Filter {
	return &Filter{
		OP:  database.NULL,
		Key: fh.key,
	}
}

func (fh *filterHelper[T]) IsNotNull() *Filter {
	return &Filter{
		OP:  database.NNULL,
		Key: fh.key,
	}
}

func (ft *Filter) And(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:    database.AND,
		Value: v,
	}
}

func (ft *Filter) Or(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:    database.OR,
		Value: v,
	}
}

func (ft *Filter) Encode() (*database.Filter, error) {
	if ft == nil {
		return nil, nil
	}
	if ft.OP == database.AND || ft.OP == database.OR {
		var values []*database.Filter
		if value, ok := ft.Value.([]*Filter); ok {
			for _, v := range value {
				if v, err := v.Encode(); err != nil {
					return nil, err
				} else {
					values = append(values, v)
				}
			}
		}
		return &database.Filter{OP: database.AND, Value: values}, nil
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

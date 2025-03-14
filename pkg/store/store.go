package store

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
	"strings"
)

type Store interface {
	Index(ctx context.Context, fields []types.String, opts ...IndexOptions) error
	Unindex(ctx context.Context, fields []types.String) error

	Insert(ctx context.Context, docs ...types.Map) error
	Remove(ctx context.Context, query types.Map) (int, error)
	Find(ctx context.Context, query types.Map) ([]types.Map, error)
}

type IndexOptions struct {
	Unique bool
	Filter types.Map
}

type store struct {
	section *Section
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
)

var _ Store = (*store)(nil)

func New() Store {
	return &store{section: NewSection()}
}

func (s *store) Index(_ context.Context, fields []types.String, opts ...IndexOptions) error {
	var unique bool
	var filter func(types.Map) bool
	for _, opt := range opts {
		if opt.Unique {
			unique = true
		}
		if opt.Filter != nil {
			filter = func(doc types.Map) bool {
				ok, err := s.match(doc, opt.Filter)
				if err != nil {
					return false
				}
				return ok
			}
		}
	}
	return s.section.Index(fields, WithUnique(unique), WithFilter(filter))
}

func (s *store) Unindex(_ context.Context, fields []types.String) error {
	return s.section.Unindex(fields)
}

func (s *store) Insert(_ context.Context, docs ...types.Map) error {
	for _, doc := range docs {
		if err := s.section.Store(doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Remove(ctx context.Context, query types.Map) (int, error) {
	docs, err := s.Find(ctx, query)
	if err != nil {
		return 0, err
	}

	for _, doc := range docs {
		if err := s.section.Delete(doc.Get(primaryKey)); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
}

func (s *store) Find(_ context.Context, query types.Map) ([]types.Map, error) {
	docs := make([]types.Map, 0)
	for _, doc := range s.section.Range() {
		if query == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := s.match(doc, query)
		if err != nil {
			return nil, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}
	return docs, nil
}

func (s *store) match(doc, query types.Value) (bool, error) {
	q, ok := query.(types.Map)
	if !ok {
		return types.Equal(doc, query), nil
	}

	for field, cond := range q.Range() {
		key, ok := field.(types.String)
		if !ok {
			return false, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", field.Interface())
		}

		if !strings.HasPrefix(key.String(), "$") {
			d, ok := doc.(types.Map)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
			}

			ok, err := s.match(d.Get(key), cond)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
			continue
		}

		switch key.String() {
		case "$eq":
			if !types.Equal(doc, cond) {
				return false, nil
			}
		case "$ne":
			if types.Equal(doc, cond) {
				return false, nil
			}
		case "$gt":
			if types.Compare(doc, cond) <= 0 {
				return false, nil
			}
		case "$lt":
			if types.Compare(doc, cond) >= 0 {
				return false, nil
			}
		case "$gte":
			if types.Compare(doc, cond) < 0 {
				return false, nil
			}
		case "$lte":
			if types.Compare(doc, cond) > 0 {
				return false, nil
			}
		default:
			return false, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}
	return true, nil
}

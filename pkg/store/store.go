package store

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
	"strings"
	"sync"
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
	indexes [][]types.String
	mu      sync.RWMutex
}

type executionPlan struct {
	key  types.String
	min  types.Value
	max  types.Value
	next *executionPlan
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
)

var _ Store = (*store)(nil)

func New() Store {
	return &store{
		section: NewSection(),
		indexes: [][]types.String{{primaryKey}},
	}
}

func (s *store) Index(_ context.Context, fields []types.String, opts ...IndexOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	if err := s.section.Index(fields, WithUnique(unique), WithFilter(filter)); err != nil {
		return err
	}

	s.indexes = append(s.indexes, fields)
	return nil
}

func (s *store) Unindex(_ context.Context, fields []types.String) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.section.Unindex(fields); err != nil {
		return err
	}

	for i := 0; i < len(s.indexes); i++ {
		idx := s.indexes[i]

		if len(fields) != len(idx) {
			continue
		}
		for j := 0; j < len(fields); j++ {
			if !types.Equal(fields[j], idx[j]) {
				continue
			}
		}

		s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		break
	}
	return nil
}

func (s *store) Insert(_ context.Context, docs ...types.Map) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, doc := range docs {
		if err := s.section.Store(doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Remove(_ context.Context, query types.Map) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	docs := make([]types.Map, 0)

	plan, err := s.explain(query)
	if err != nil {
		return 0, err
	}

	c := Cursor(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	for _, doc := range c.Range() {
		if query == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := s.match(doc, query)
		if err != nil {
			return 0, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	for _, doc := range docs {
		if err := s.section.Delete(doc.Get(primaryKey)); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
}

func (s *store) Find(_ context.Context, query types.Map) ([]types.Map, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs := make([]types.Map, 0)

	plan, err := s.explain(query)
	if err != nil {
		return nil, err
	}

	c := Cursor(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	for _, doc := range c.Range() {
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

func (s *store) explain(query types.Value) (*executionPlan, error) {
	q, ok := query.(types.Map)
	if !ok {
		return nil, nil
	}

	var plans []*executionPlan
	for _, idx := range s.indexes {
		plan := &executionPlan{}
		curr := plan

		for _, field := range idx {
			cond := q.Get(field)
			if cond == nil {
				continue
			}

			next := &executionPlan{key: field}
			if c, ok := cond.(types.Map); !ok {
				next.min = cond
				next.max = cond
			} else if v := c.Get(types.NewString("$eq")); v != nil {
				next.min = v
				next.max = v
			} else if v := c.Get(types.NewString("$gt")); v != nil {
				next.min = v
			} else if v := c.Get(types.NewString("$gte")); v != nil {
				next.min = v
			} else if v := c.Get(types.NewString("$lt")); v != nil {
				next.max = v
			} else if v := c.Get(types.NewString("$lte")); v != nil {
				next.max = v
			} else {
				break
			}

			curr.next = next
			curr = next
		}

		if plan.next != nil {
			plans = append(plans, plan.next)
		}
	}

	var plan *executionPlan
	for _, p := range plans {
		if plan == nil || p.cost() < plan.cost() {
			plan = p
		}
	}
	return plan, nil
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

func (e *executionPlan) cost() int {
	if e.next != nil {
		return 1 + e.next.cost()
	}
	return 1
}

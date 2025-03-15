package store

import (
	"context"
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
)

type Store interface {
	Watch(ctx context.Context, filter any) (Stream, error)

	Index(ctx context.Context, keys []string, opts ...IndexOptions) error
	Unindex(ctx context.Context, keys []string) error

	Insert(ctx context.Context, docs []any, opts ...InsertOptions) error
	Update(ctx context.Context, filter, update any, opts ...UpdateOptions) (int, error)
	Delete(ctx context.Context, filter any, opts ...DeleteOptions) (int, error)
	Find(ctx context.Context, filter any, opts ...FindOptions) (Cursor, error)
}

type IndexOptions struct {
	Unique bool
	Filter types.Map
}

type InsertOptions struct {
}

type DeleteOptions struct {
}

type UpdateOptions struct {
	Upsert bool
}

type FindOptions struct {
	Limit int
	Sort  types.Map
}

type store struct {
	section *section
	indexes [][]types.String
	streams []*stream
	mu      sync.RWMutex
}

type executionPlan struct {
	key  types.String
	min  types.Value
	max  types.Value
	next *executionPlan
}

var (
	ErrKeyMissing   = errors.New("key is missing")
	ErrKeyDuplicate = errors.New("key already exists")
	ErrKeyNotFound  = errors.New("key not found")

	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrUnsupportedType      = errors.New("unsupported type")
)

var _ Store = (*store)(nil)

func New() Store {
	return &store{
		section: newSection(),
		indexes: [][]types.String{{types.NewString("id")}},
	}
}

func (s *store) Watch(ctx context.Context, filter any) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var f types.Map
	if filter != nil {
		var err error
		if f, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return nil, err
		}
	}

	strm := newStream(f)
	s.streams = append(s.streams, strm)

	if ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				_ = strm.Close(ctx)
			case <-strm.Done():
			}
		}()
	}

	go func() {
		<-strm.Done()

		s.mu.Lock()
		defer s.mu.Unlock()

		for i := 0; i < len(s.streams); i++ {
			if s.streams[i] == strm {
				s.streams = append(s.streams[:i], s.streams[i+1:]...)
				break
			}
		}
	}()

	return strm, nil
}

func (s *store) Index(_ context.Context, keys []string, opts ...IndexOptions) error {
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
				ok, err := match(doc, opt.Filter)
				if err != nil {
					return false
				}
				return ok
			}
		}
	}

	idx := make([]types.String, 0, len(keys))
	for _, k := range keys {
		idx = append(idx, types.NewString(k))
	}

	if err := s.section.Index(idx, withUnique(unique), withFilter(filter)); err != nil {
		return err
	}

	s.indexes = append(s.indexes, idx)
	return nil
}

func (s *store) Unindex(_ context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := make([]types.String, 0, len(keys))
	for _, k := range keys {
		idx = append(idx, types.NewString(k))
	}

	if err := s.section.Unindex(idx); err != nil {
		return err
	}

	for i := 0; i < len(s.indexes); i++ {
		if len(s.indexes[i]) != len(idx) {
			continue
		}
		for j := 0; j < len(idx); j++ {
			if !types.Equal(s.indexes[i][j], idx[j]) {
				continue
			}
		}

		s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		break
	}
	return nil
}

func (s *store) Insert(_ context.Context, docs []any, _ ...InsertOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, doc := range docs {
		val, err := types.Cast[types.Map](types.Marshal(doc))
		if err != nil {
			return err
		}
		if err := s.section.Store(val); err != nil {
			return err
		}
		if err := s.emit(types.NewString("insert"), val); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Update(_ context.Context, filter, update any, opts ...UpdateOptions) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var upsert bool
	for _, opt := range opts {
		if opt.Upsert {
			upsert = opt.Upsert
		}
	}

	var f types.Map
	if filter != nil {
		var err error
		if f, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return 0, err
		}
	}

	u, err := types.Cast[types.Map](types.Marshal(update))
	if err != nil {
		return 0, err
	}

	plan, err := s.explain(f)
	if err != nil {
		return 0, err
	}

	c := scanner(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	docs := make([]types.Map, 0)
	for _, doc := range c.Range() {
		if f == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := match(doc, f)
		if err != nil {
			return 0, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	if upsert && len(docs) == 0 {
		doc, err := types.Cast[types.Map](apply(f))
		if err != nil {
			return 0, err
		}

		doc, err = patch(doc, u)
		if err != nil {
			return 0, err
		}

		if err := s.section.Store(doc); err != nil {
			return 0, err
		}
		if err := s.emit(types.NewString("insert"), doc); err != nil {
			return 0, err
		}
		return 1, nil
	}

	for i := 0; i < len(docs); i++ {
		doc, err := patch(docs[i], u)
		if err != nil {
			return 0, err
		}
		docs[i] = doc
	}

	for _, doc := range docs {
		if err := s.section.Swap(doc); err != nil {
			return 0, err
		}
		if err := s.emit(types.NewString("update"), doc); err != nil {
			return 0, err
		}
	}

	return len(docs), nil
}

func (s *store) Delete(_ context.Context, filter any, _ ...DeleteOptions) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var f types.Map
	if filter != nil {
		var err error
		if f, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return 0, err
		}
	}

	plan, err := s.explain(f)
	if err != nil {
		return 0, err
	}

	c := scanner(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	docs := make([]types.Map, 0)
	for _, doc := range c.Range() {
		if f == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := match(doc, f)
		if err != nil {
			return 0, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	for _, doc := range docs {
		if err := s.section.Delete(doc.Get(types.NewString("id"))); err != nil {
			return 0, err
		}
		if err := s.emit(types.NewString("delete"), doc); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
}

func (s *store) Find(_ context.Context, filter any, opts ...FindOptions) (Cursor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var limit int
	var sort types.Map
	for _, opt := range opts {
		if opt.Limit != 0 {
			limit = opt.Limit
		}
		if opt.Sort != nil {
			sort = opt.Sort
		}
	}

	var f types.Map
	if filter != nil {
		var err error
		if f, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return nil, err
		}
	}

	plan, err := s.explain(f)
	if err != nil {
		return nil, err
	}

	c := scanner(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	docs := make([]types.Map, 0)
	for _, doc := range c.Range() {
		if f == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := match(doc, f)
		if err != nil {
			return nil, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	if sort != nil {
		slices.SortFunc(docs, func(x, y types.Map) int {
			for field, o := range sort.Range() {
				val1 := x.Get(field)
				val2 := y.Get(field)

				order := 1
				_ = types.Unmarshal(o, &order)

				if comp := types.Compare(val1, val2); comp != 0 {
					return comp * order
				}
			}
			return 0
		})
	}

	if limit > 0 && len(docs) > limit {
		docs = docs[:limit]
	}

	return newCursor(docs), nil
}

func (s *store) explain(filter types.Value) (*executionPlan, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return nil, nil
	}

	var plans []*executionPlan
	for _, idx := range s.indexes {
		plan := &executionPlan{}
		curr := plan

		for _, key := range idx {
			value := f.Get(key)
			if value == nil {
				continue
			}

			next := &executionPlan{key: key}
			if val, ok := value.(types.Map); ok {
				if v := val.Get(types.NewString("$eq")); v != nil {
					next.min = v
					next.max = v
				} else {
					var lower types.Value
					var lowers []types.Value
					if v := val.Get(types.NewString("$gt")); v != nil {
						lowers = append(lowers, v)
					}
					if v := val.Get(types.NewString("$gte")); v != nil {
						lowers = append(lowers, v)
					}
					if len(lowers) > 0 {
						lower = lowers[0]
						for i := 1; i < len(lowers); i++ {
							if types.Compare(lowers[i], lower) > 0 {
								lower = lowers[i]
							}
						}
						next.min = lower
					}

					var upper types.Value
					var uppers []types.Value
					if v := val.Get(types.NewString("$lt")); v != nil {
						uppers = append(uppers, v)
					}
					if v := val.Get(types.NewString("$lte")); v != nil {
						uppers = append(uppers, v)
					}
					if len(uppers) > 0 {
						upper = uppers[0]
						for i := 1; i < len(uppers); i++ {
							if types.Compare(uppers[i], upper) < 0 {
								upper = uppers[i]
							}
						}
						next.max = upper
					}
				}
			}

			if next.min == nil && next.max == nil {
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

func (s *store) emit(op types.String, doc types.Map) error {
	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessagef(ErrKeyMissing, "key: %s", "id")
	}

	for _, strm := range s.streams {
		ok, err := strm.Match(doc)
		if err != nil {
			return err
		}
		if ok {
			strm.Emit(types.NewMap(types.NewString("op"), op, types.NewString("id"), id))
		}
	}
	return nil
}

func (e *executionPlan) cost() int {
	if e.next != nil {
		return 1 + e.next.cost()
	}
	return 1
}

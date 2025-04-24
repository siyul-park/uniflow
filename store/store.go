package store

import (
	"context"
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/types"
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
	Filter any
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
	Skip  int
	Sort  any
}

type store struct {
	segment *segment
	streams []*stream
	mu      sync.RWMutex
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
	return &store{segment: newSegment()}
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
			val, err := types.Cast[types.Map](types.Marshal(opt.Filter))
			if err != nil {
				return err
			}
			filter = func(doc types.Map) bool {
				ok, err := match(doc, val)
				if err != nil {
					return false
				}
				return ok
			}
		}
	}

	idx := &index{Keys: make([]types.String, 0, len(keys)), Unique: unique, Filter: filter}
	for _, k := range keys {
		idx.Keys = append(idx.Keys, types.NewString(k))
	}

	for _, i := range s.segment.Indexes() {
		if slices.Equal(i.Keys, idx.Keys) {
			if err := s.segment.Unindex(i); err != nil {
				return err
			}
		}
	}
	return s.segment.Index(idx)
}

func (s *store) Unindex(_ context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := &index{Keys: make([]types.String, 0, len(keys))}
	for _, k := range keys {
		idx.Keys = append(idx.Keys, types.NewString(k))
	}

	for _, i := range s.segment.Indexes() {
		if slices.Equal(i.Keys, idx.Keys) {
			if err := s.segment.Unindex(i); err != nil {
				return err
			}
		}
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
		if err := s.segment.Store(val); err != nil {
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

	docs, err := s.find(f)
	if err != nil {
		return 0, err
	}

	if upsert && len(docs) == 0 {
		doc, err := types.Cast[types.Map](extract(f))
		if err != nil {
			return 0, err
		}

		doc, err = patch(doc, u)
		if err != nil {
			return 0, err
		}

		if err := s.segment.Store(doc); err != nil {
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
		if err := s.segment.Swap(doc); err != nil {
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

	docs, err := s.find(f)
	if err != nil {
		return 0, err
	}

	for _, doc := range docs {
		if err := s.segment.Delete(doc.Get(types.NewString("id"))); err != nil {
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
	var skip int
	var sort types.Map
	for _, opt := range opts {
		if opt.Limit > 0 {
			limit = opt.Limit
		}
		if opt.Skip > 0 {
			skip = opt.Skip
		}
		if opt.Sort != nil {
			var err error
			if sort, err = types.Cast[types.Map](types.Marshal(opt.Sort)); err != nil {
				return nil, err
			}
		}
	}

	var f types.Map
	if filter != nil {
		var err error
		if f, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return nil, err
		}
	}

	docs, err := s.find(f)
	if err != nil {
		return nil, err
	}

	if sort != nil {
		slices.SortFunc(docs, func(x, y types.Map) int {
			for field, o := range sort.Range() {
				val1 := x.Get(field)
				val2 := y.Get(field)

				if comp := types.Compare(val1, val2); comp != 0 {
					order := 1
					_ = types.Unmarshal(o, &order)
					return comp * order
				}
			}
			return 0
		})
	}

	if skip > len(docs) {
		skip = len(docs)
	}
	if limit == 0 {
		limit = len(docs)
	}

	limit = skip + limit
	if limit > len(docs) {
		limit = len(docs)
	}

	docs = docs[skip:limit]

	return newCursor(docs), nil
}

func (s *store) find(filter types.Map) ([]types.Map, error) {
	plan, err := s.explain(filter)
	if err != nil {
		return nil, err
	}

	scan := scanner(s.segment)
	for plan != nil {
		scan = scan.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	var docs []types.Map
	for _, doc := range scan.Range() {
		if filter == nil {
			docs = append(docs, doc)
			continue
		}

		if ok, err := match(doc, filter); err != nil {
			return nil, err
		} else if ok {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (s *store) explain(filter types.Value) (*executionPlan, error) {
	if filter == nil {
		return nil, nil
	}

	doc, _ := types.Cast[types.Map](extract(filter))

	var plans []*executionPlan
	for _, idx := range s.segment.Indexes() {
		if idx.Filter != nil && (doc == nil || !idx.Filter(doc)) {
			continue
		}
		if plan := newExecutionPlan(idx.Keys, filter); plan != nil {
			plans = append(plans, plan)
		}
	}

	var plan *executionPlan
	for _, p := range plans {
		if plan == nil || p.lenght() > plan.lenght() {
			plan = p
		}
	}
	return plan, nil
}

func (s *store) emit(op types.String, doc types.Map) error {
	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessage(ErrKeyMissing, "key: id")
	}

	for _, strm := range s.streams {
		if ok, err := strm.Match(doc); err != nil {
			return err
		} else if ok {
			strm.Emit(types.NewMap(types.NewString("op"), op, types.NewString("id"), id))
		}
	}
	return nil
}

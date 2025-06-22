package driver

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Store defines the interface for a basic document store.
type Store interface {
	Watch(ctx context.Context, filter any) (Stream, error)

	Indexes(ctx context.Context) ([][]string, error)
	Index(ctx context.Context, keys []string, opts ...IndexOptions) error
	Unindex(ctx context.Context, keys []string) error

	Insert(ctx context.Context, docs []any, opts ...InsertOptions) error
	Update(ctx context.Context, filter, update any, opts ...UpdateOptions) (int, error)
	Delete(ctx context.Context, filter any, opts ...DeleteOptions) (int, error)
	Find(ctx context.Context, filter any, opts ...FindOptions) (Cursor, error)
}

// IndexOptions represents options when creating an index.
type IndexOptions struct {
	Unique bool
	Filter any
}

// InsertOptions represents options when inserting documents.
type InsertOptions struct {
}

// DeleteOptions represents options when deleting documents.
type DeleteOptions struct {
}

// UpdateOptions represents options when updating documents.
type UpdateOptions struct {
	Upsert bool
}

// FindOptions represents options when finding documents.
type FindOptions struct {
	Limit int
	Skip  int
	Sort  any
}

type store struct {
	segment *segment
	streams []*stream
	filters []types.Map
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

// NewStore creates and returns a new in-memory store instance.
func NewStore() Store {
	return &store{segment: newSegment()}
}

func (s *store) Watch(ctx context.Context, filter any) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var fltr types.Map
	if filter != nil {
		var err error
		if fltr, err = types.Cast[types.Map](types.Marshal(filter)); err != nil {
			return nil, err
		}
	}

	strm := newStream()

	s.streams = append(s.streams, strm)
	s.filters = append(s.filters, fltr)

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
				s.filters = append(s.filters[:i], s.filters[i+1:]...)
				break
			}
		}
	}()

	return strm, nil
}

func (s *store) Indexes(_ context.Context) ([][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	indexes := s.segment.Indexes()
	keys := make([][]string, 0, len(indexes))
	for _, idx := range indexes {
		keys = append(keys, lo.Map(idx.Keys, func(item types.String, index int) string {
			return item.String()
		}))
	}
	return keys, nil
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
				ok, err := s.match(doc, val)
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
		doc, err := types.Cast[types.Map](s.extract(f))
		if err != nil {
			return 0, err
		}

		doc, err = s.patch(doc, u)
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
		doc, err := s.patch(docs[i], u)
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

		if ok, err := s.match(doc, filter); err != nil {
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

	doc, _ := types.Cast[types.Map](s.extract(filter))

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

	for i, strm := range s.streams {
		if filter := s.filters[i]; filter != nil {
			if ok, err := s.match(doc, filter); err != nil {
				return err
			} else if !ok {
				continue
			}
		}
		strm.Emit(types.NewMap(types.NewString("op"), op, types.NewString("id"), id))
	}
	return nil
}

func (s *store) match(doc, filter types.Value) (bool, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return types.Equal(doc, filter), nil
	}

	for k, value := range f.Range() {
		key, ok := k.(types.String)
		if !ok {
			return false, errors.WithMessagef(ErrUnsupportedType, "key: %v", k.Interface())
		}

		if !strings.HasPrefix(key.String(), "$") {
			d, ok := doc.(types.Map)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "doc: %v", doc.Interface())
			}

			ok, err := s.match(d.Get(key), value)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
			continue
		}

		switch key.String() {
		case "$exists":
			if reflect.ValueOf(value).IsZero() {
				return value == nil, nil
			}
			return value != nil, nil
		case "$eq":
			if !types.Equal(doc, value) {
				return false, nil
			}
		case "$ne":
			if types.Equal(doc, value) {
				return false, nil
			}
		case "$gt":
			if types.Compare(doc, value) <= 0 {
				return false, nil
			}
		case "$lt":
			if types.Compare(doc, value) >= 0 {
				return false, nil
			}
		case "$gte":
			if types.Compare(doc, value) < 0 {
				return false, nil
			}
		case "$lte":
			if types.Compare(doc, value) > 0 {
				return false, nil
			}
		case "$and":
			vals, ok := value.(types.Slice)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				match, err := s.match(doc, sub)
				if err != nil {
					return false, err
				}
				if !match {
					return false, nil
				}
			}
		case "$or":
			vals, ok := value.(types.Slice)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				match, err := s.match(doc, sub)
				if err != nil {
					return false, err
				}
				if match {
					return true, nil
				}
			}
			return false, nil
		default:
			return false, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}
	return true, nil
}

func (s *store) patch(doc, update types.Map) (types.Map, error) {
	doc = doc.Mutable()
	for k, value := range update.Range() {
		key, ok := k.(types.String)
		if !ok {
			return nil, errors.WithMessagef(ErrUnsupportedType, "key: %v", k.Interface())
		}

		switch key.String() {
		case "$set":
			val, ok := value.(types.Map)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for k, v := range val.Range() {
				doc.Set(k, v)
			}
		case "$unset":
			val, ok := value.(types.Map)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for k := range val.Range() {
				doc.Delete(k)
			}
		default:
			return nil, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}
	return doc.Immutable(), nil
}

func (s *store) extract(filter types.Value) (types.Value, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return filter, nil
	}

	doc := types.NewMap().Mutable()

	for k, value := range f.Range() {
		key, ok := k.(types.String)
		if !ok {
			continue
		}

		if !strings.HasPrefix(key.String(), "$") {
			child, err := s.extract(value)
			if err != nil {
				return nil, err
			}
			if child != nil {
				doc = doc.Set(key, child)
			}
			continue
		}

		switch key.String() {
		case "$eq":
			return value, nil
		case "$and", "$or":
			vals, ok := value.(types.Slice)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				child, err := types.Cast[types.Map](s.extract(sub))
				if err != nil {
					return nil, err
				}

				for key, val := range child.Range() {
					if doc.Has(key) {
						return nil, errors.WithMessagef(ErrKeyDuplicate, "key: %v", key.Interface())
					}
					doc = doc.Set(key, val)
				}
			}
		default:
			return nil, nil
		}
	}

	return doc.Immutable(), nil
}

package store

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
)

type Store interface {
	Watch(ctx context.Context, filter types.Map) (Stream, error)

	Index(ctx context.Context, keys []types.String, opts ...IndexOptions) error
	Unindex(ctx context.Context, keys []types.String) error

	Insert(ctx context.Context, docs []types.Map, opts ...InsertOptions) error
	Update(ctx context.Context, filter, update types.Map, opts ...UpdateOptions) (int, error)
	Delete(ctx context.Context, filter types.Map, opts ...DeleteOptions) (int, error)
	Find(ctx context.Context, filter types.Map, opts ...FindOptions) ([]types.Map, error)
}

type Stream interface {
	Next() <-chan types.Map
	Done() <-chan struct{}
	Close() error
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
	section *Section
	indexes [][]types.String
	stream  []*stream
	filters []types.Map
	mu      sync.RWMutex
}

type stream struct {
	in   chan types.Map
	out  chan types.Map
	done chan struct{}
	mu   sync.Mutex
}

type executionPlan struct {
	key  types.String
	min  types.Value
	max  types.Value
	next *executionPlan
}

const (
	KeyID = "id"
	KeyOP = "op"
)

var (
	ErrKeyMissing   = errors.New("key is missing")
	ErrKeyDuplicate = errors.New("key already exists")
	ErrKeyNotFound  = errors.New("key not found")

	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrUnsupportedType      = errors.New("unsupported type")
)

var _ Store = (*store)(nil)
var _ Stream = (*stream)(nil)

func New() Store {
	return &store{
		section: NewSection(),
		indexes: [][]types.String{{types.NewString(KeyID)}},
	}
}

func (s *store) Watch(ctx context.Context, filter types.Map) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	strm := &stream{
		in:   make(chan types.Map),
		out:  make(chan types.Map),
		done: make(chan struct{}),
	}

	go func() {
		defer close(strm.out)
		defer close(strm.in)

		buffer := make([]types.Map, 0, 2)
		for {
			var event types.Map
			select {
			case event = <-strm.in:
			case <-strm.done:
				return
			}

			select {
			case strm.out <- event:
			default:
				buffer = append(buffer, event)

				for len(buffer) > 0 {
					select {
					case event = <-strm.in:
						buffer = append(buffer, event)
					case strm.out <- buffer[0]:
						buffer = buffer[1:]
					}
				}
			}
		}
	}()

	if ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				_ = strm.Close()
			case <-strm.Done():
			}
		}()
	}

	go func() {
		<-strm.Done()

		s.mu.Lock()
		defer s.mu.Unlock()

		for i := 0; i < len(s.stream); i++ {
			if s.stream[i] == strm {
				s.stream = append(s.stream[:i], s.stream[i+1:]...)
				s.filters = append(s.filters[:i], s.filters[i+1:]...)
				break
			}
		}
	}()

	s.stream = append(s.stream, strm)
	s.filters = append(s.filters, filter)

	return strm, nil
}

func (s *store) Index(_ context.Context, keys []types.String, opts ...IndexOptions) error {
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

	if err := s.section.Index(keys, WithUnique(unique), WithFilter(filter)); err != nil {
		return err
	}

	s.indexes = append(s.indexes, keys)
	return nil
}

func (s *store) Unindex(_ context.Context, keys []types.String) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.section.Unindex(keys); err != nil {
		return err
	}

	for i := 0; i < len(s.indexes); i++ {
		idx := s.indexes[i]

		if len(keys) != len(idx) {
			continue
		}
		for j := 0; j < len(keys); j++ {
			if !types.Equal(keys[j], idx[j]) {
				continue
			}
		}

		s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		break
	}
	return nil
}

func (s *store) Insert(_ context.Context, docs []types.Map, _ ...InsertOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, doc := range docs {
		if err := s.section.Store(doc); err != nil {
			return err
		}
		if err := s.emit(types.NewString("insert"), doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Update(_ context.Context, filter, update types.Map, opts ...UpdateOptions) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var upsert bool
	for _, opt := range opts {
		if opt.Upsert {
			upsert = opt.Upsert
		}
	}

	docs := make([]types.Map, 0)

	plan, err := s.explain(filter)
	if err != nil {
		return 0, err
	}

	c := Cursor(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	for _, doc := range c.Range() {
		if filter == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := s.match(doc, filter)
		if err != nil {
			return 0, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	if upsert && len(docs) == 0 {
		d, err := s.apply(filter)
		if err != nil {
			return 0, err
		}
		doc, ok := d.(types.Map)
		if !ok {
			return 0, errors.WithMessagef(ErrUnsupportedType, "value: %v", d)
		}

		doc, err = s.patch(doc, update)
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
		doc, err := s.patch(docs[i], update)
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

func (s *store) Delete(_ context.Context, filter types.Map, _ ...DeleteOptions) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	docs := make([]types.Map, 0)

	plan, err := s.explain(filter)
	if err != nil {
		return 0, err
	}

	c := Cursor(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	for _, doc := range c.Range() {
		if filter == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := s.match(doc, filter)
		if err != nil {
			return 0, err
		}
		if ok {
			docs = append(docs, doc.(types.Map))
		}
	}

	for _, doc := range docs {
		if err := s.section.Delete(doc.Get(types.NewString(KeyID))); err != nil {
			return 0, err
		}
		if err := s.emit(types.NewString("delete"), doc); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
}

func (s *store) Find(_ context.Context, filter types.Map, opts ...FindOptions) ([]types.Map, error) {
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

	docs := make([]types.Map, 0)

	plan, err := s.explain(filter)
	if err != nil {
		return nil, err
	}

	c := Cursor(s.section)
	for plan != nil {
		c = c.Scan(plan.key, plan.min, plan.max)
		plan = plan.next
	}

	for _, doc := range c.Range() {
		if filter == nil {
			docs = append(docs, doc.(types.Map))
			continue
		}
		ok, err := s.match(doc, filter)
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
	return docs, nil
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

func (s *store) match(doc, filter types.Value) (bool, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return false, errors.WithMessagef(ErrUnsupportedType, "filter: %v", filter.Interface())
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

func (s *store) apply(filter types.Value) (types.Value, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return nil, errors.WithMessagef(ErrUnsupportedType, "filter: %v", filter.Interface())
	}

	doc := types.NewMap().Mutable()

	for k, value := range f.Range() {
		key, ok := k.(types.String)
		if !ok {
			continue
		}

		if !strings.HasPrefix(key.String(), "$") {
			child, err := s.apply(value)
			if err != nil {
				return nil, err
			}
			doc = doc.Set(key, child)
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
				c, err := s.apply(sub)
				if err != nil {
					return nil, err
				}

				child, ok := c.(types.Map)
				if !ok {
					return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", child.Interface())
				}

				for key, val := range child.Range() {
					if doc.Has(key) {
						return nil, errors.WithMessagef(ErrUnsupportedOperation, "key: %v", key.Interface())
					}
					doc = doc.Set(key, val)
				}
			}
		default:
			return nil, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}

	return doc.Immutable(), nil
}

func (s *store) emit(op types.String, doc types.Map) error {
	id := doc.Get(types.NewString(KeyID))
	if id == nil {
		return errors.WithMessagef(ErrKeyMissing, "key: %s", KeyID)
	}

	for i := 0; i < len(s.stream); i++ {
		ok := true
		if s.filters[i] != nil {
			var err error
			ok, err = s.match(doc, s.filters[i])
			if err != nil {
				return err
			}
		}
		if ok {
			s.stream[i].in <- types.NewMap(types.NewString(KeyOP), op, types.NewString(KeyID), id)
		}
	}
	return nil
}

func (s *stream) Next() <-chan types.Map {
	return s.out
}

func (s *stream) Done() <-chan struct{} {
	return s.done
}

func (s *stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	default:
		close(s.done)
		return nil
	}
}

func (e *executionPlan) cost() int {
	if e.next != nil {
		return 1 + e.next.cost()
	}
	return 1
}

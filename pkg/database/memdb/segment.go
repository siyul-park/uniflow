package memdb

import (
	"sort"
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Segment struct {
	data    maps.Map
	indexes []maps.Map
	models  []database.IndexModel
	mu      sync.RWMutex
}

func newSegment() *Segment {
	s := &Segment{data: treemap.NewWith(comparator)}

	primary := database.IndexModel{
		Keys:    []string{"id"},
		Name:    "_id",
		Unique:  true,
		Partial: nil,
	}

	s.models = append(s.models, primary)
	s.indexes = append(s.indexes, treemap.NewWith(comparator))

	return s
}

func (s *Segment) Models() ([]database.IndexModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.models, nil
}

func (s *Segment) Index(index database.IndexModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, model := range s.models {
		if model.Name == index.Name {
			s.models = append(s.models[:i], s.models[i+1:]...)
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		}
	}

	s.models = append(s.models, index)
	s.indexes = append(s.indexes, treemap.NewWith(comparator))

	for _, doc := range s.data.Values() {
		if err := s.index(doc.(*primitive.Map)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Segment) UnIndex(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, model := range s.models {
		if model.Name == name {
			s.models = append(s.models[:i], s.models[i+1:]...)
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		}
	}

	return nil
}

func (s *Segment) Find(filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	opt := database.MergeFindOptions(opts)

	limit := -1
	skip := 0
	var sorts []database.Sort

	if opt != nil {
		if opt.Limit != nil {
			limit = lo.FromPtr(opt.Limit)
		}
		if opt.Skip != nil {
			skip = lo.FromPtr(opt.Skip)
		}
		if opt.Sorts != nil {
			sorts = opt.Sorts
		}
	}

	match := parseFilter(filter)

	var docs []*primitive.Map
	for _, value := range s.data.Values() {
		if len(sorts) == 0 && limit >= 0 && len(docs) == limit+skip {
			continue
		}
		if match(value.(*primitive.Map)) {
			docs = append(docs, value.(*primitive.Map))
		}
	}

	if skip >= len(docs) {
		return nil, nil
	}
	if len(sorts) > 0 {
		compare := parseSorts(sorts)
		sort.Slice(docs, func(i, j int) bool {
			return compare(docs[i], docs[j])
		})
	}
	docs = docs[skip:]
	if limit >= 0 && len(docs) > limit {
		docs = docs[:limit]
	}

	return docs, nil
}

func (s *Segment) Insert(docs []*primitive.Map) ([]primitive.Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]primitive.Value, len(docs))
	for i, doc := range docs {
		if id, ok := doc.Get(keyID); !ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
		} else {
			ids[i] = id
		}
	}

	for _, id := range ids {
		if _, ok := s.data.Get(id); ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKDuplicated), database.ErrCodeWrite)
		}
	}

	for i, doc := range docs {
		if err := s.index(doc); err != nil {
			for i--; i >= 0; i-- {
				s.data.Remove(ids[i])
				s.unindex(doc)
			}
			return nil, err
		}
		s.data.Put(ids[i], doc)
	}

	return ids, nil
}

func (s *Segment) Delete(docs []*primitive.Map) ([]*primitive.Map, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	docs = lo.Filter[*primitive.Map](docs, func(item *primitive.Map, _ int) bool {
		return item != nil && item.GetOr(keyID, nil) != nil
	})

	for _, doc := range docs {
		s.unindex(doc)
		s.data.Remove(doc.GetOr(keyID, nil))
	}

	return docs, nil
}

func (s *Segment) Drop() []*primitive.Map {
	s.mu.Lock()
	defer s.mu.Unlock()

	var data []*primitive.Map
	for _, doc := range s.data.Values() {
		data = append(data, doc.(*primitive.Map))
	}

	s.data = treemap.NewWith(comparator)
	for i := range s.indexes {
		s.indexes[i] = treemap.NewWith(comparator)
	}

	return data
}

func (s *Segment) index(doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrIndexConflict)
	}

	for i, model := range s.models {
		match := parseFilter(model.Partial)
		if !match(doc) {
			continue
		}

		cur := s.indexes[i]

		for i, k := range model.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(maps.Map)
			if !ok {
				child = treemap.NewWith(comparator)
				cur.Put(value, child)
			}

			if i < len(model.Keys)-1 {
				cur = child
			} else {
				child.Put(id, nil)

				if model.Unique && child.Size() > 1 {
					child.Remove(id)
					s.unindex(doc)
					return errors.WithStack(ErrIndexConflict)
				}
			}
		}
	}

	return nil
}

func (s *Segment) unindex(doc *primitive.Map) {
	id, ok := doc.Get(keyID)
	if !ok {
		return
	}

	for i, model := range s.models {
		cur := s.indexes[i]
		nodes := []maps.Map{cur}
		keys := []primitive.Value{nil}

		for i, k := range model.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(maps.Map)
			if !ok {
				nodes = nil
				keys = nil

				break
			}

			nodes = append(nodes, child)
			keys = append(keys, value)

			if i < len(model.Keys)-1 {
				cur = child
			} else {
				child.Remove(id)
			}
		}

		for i := len(nodes) - 1; i >= 0; i-- {
			child := nodes[i]

			if child.Empty() && i > 0 {
				parent := nodes[i-1]
				key := keys[i]

				parent.Remove(key)
			}
		}
	}
}

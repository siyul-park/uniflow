package memdb

import (
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Segment struct {
	data    maps.Map
	indexes []maps.Map
	models  []Model
	mu      sync.RWMutex
}

type Model struct {
	Name   string
	Keys   []string
	Unique bool
	Match  func(*primitive.Map) bool
}

var (
	ErrPKNotFound   = errors.New("primary key is not found")
	ErrPKDuplicated = errors.New("primary key is duplicated")
)

var keyID = primitive.NewString("id")

var comparator = utils.Comparator(func(a, b any) int {
	return primitive.Compare(a.(primitive.Value), b.(primitive.Value))
})

func newSegment() *Segment {
	s := &Segment{data: treemap.NewWith(comparator)}

	primary := Model{
		Keys:   []string{"id"},
		Name:   "_id",
		Unique: true,
		Match:  func(_ *primitive.Map) bool { return true },
	}

	s.models = append(s.models, primary)
	s.indexes = append(s.indexes, treemap.NewWith(comparator))

	return s
}

func (s *Segment) Index(index Model) error {
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

func (s *Segment) Unindex(name string) error {
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

func (s *Segment) Set(doc *primitive.Map) (primitive.Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := doc.Get(keyID)
	if !ok {
		return nil, errors.WithStack(ErrPKNotFound)
	}

	if _, ok := s.data.Get(id); ok {
		return nil, errors.WithStack(ErrPKDuplicated)
	}

	if err := s.index(doc); err != nil {
		return nil, err
	}
	s.data.Put(id, doc)

	return id, nil
}

func (s *Segment) Delete(doc *primitive.Map) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := doc.Get(keyID)
	if !ok {
		return false
	}

	if _, ok := s.data.Get(id); !ok {
		return false
	}

	s.unindex(doc)
	s.data.Remove(doc.GetOr(keyID, nil))

	return true
}

func (s *Segment) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, doc := range s.data.Values() {
		if !f(doc.(*primitive.Map)) {
			break
		}
	}
}

func (s *Segment) Drop() []*primitive.Map {
	s.mu.Lock()
	defer s.mu.Unlock()

	var data []*primitive.Map
	for _, doc := range s.data.Values() {
		data = append(data, doc.(*primitive.Map))
	}

	s.data.Clear()
	for _, index := range s.indexes {
		index.Clear()
	}

	return data
}

func (s *Segment) index(doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrPKNotFound)
	}

	for i, model := range s.models {
		if !model.Match(doc) {
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

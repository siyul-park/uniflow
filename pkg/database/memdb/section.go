package memdb

import (
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Section struct {
	data        maps.Map
	indexes     []*treemap.Map
	constraints []Constraint
	mu          *sync.RWMutex
}

type Constraint struct {
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

func newSection() *Section {
	s := &Section{
		data: treemap.NewWith(comparator),
		mu:   &sync.RWMutex{},
	}

	primary := Constraint{
		Keys:   []string{"id"},
		Name:   "_id",
		Unique: true,
		Match:  func(_ *primitive.Map) bool { return true },
	}

	s.constraints = append(s.constraints, primary)
	s.indexes = append(s.indexes, treemap.NewWith(comparator))

	return s
}

func (s *Section) AddConstraint(constraint Constraint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.constraints {
		if c.Name == constraint.Name {
			s.constraints = append(s.constraints[:i], s.constraints[i+1:]...)
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		}
	}

	s.constraints = append(s.constraints, constraint)
	s.indexes = append(s.indexes, treemap.NewWith(comparator))

	for _, doc := range s.data.Values() {
		if err := s.index(doc.(*primitive.Map)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Section) DropConstraint(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, constraint := range s.constraints {
		if constraint.Name == name {
			s.constraints = append(s.constraints[:i], s.constraints[i+1:]...)
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		}
	}

	return nil
}

func (s *Section) Set(doc *primitive.Map) (primitive.Value, error) {
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

func (s *Section) Delete(doc *primitive.Map) bool {
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

func (s *Section) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, doc := range s.data.Values() {
		if !f(doc.(*primitive.Map)) {
			break
		}
	}
}

func (s *Section) Scan(name string, min, max primitive.Value) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i, constraint := range s.constraints {
		if constraint.Name != name {
			continue
		}

		return &Sector{
			data:  s.data,
			keys:  constraint.Keys[1:],
			index: s.indexes[i],
			min:   min,
			max:   max,
			mu:    s.mu,
		}, true
	}

	return nil, false
}

func (s *Section) Drop() []*primitive.Map {
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

func (s *Section) index(doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrPKNotFound)
	}

	for i, constraint := range s.constraints {
		if !constraint.Match(doc) {
			continue
		}

		cur := s.indexes[i]

		for i, k := range constraint.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(*treemap.Map)
			if !ok {
				child = treemap.NewWith(comparator)
				cur.Put(value, child)
			}

			if i < len(constraint.Keys)-1 {
				cur = child
			} else {
				child.Put(id, nil)

				if constraint.Unique && child.Size() > 1 {
					child.Remove(id)
					s.unindex(doc)
					return errors.WithStack(ErrIndexConflict)
				}
			}
		}
	}

	return nil
}

func (s *Section) unindex(doc *primitive.Map) {
	id, ok := doc.Get(keyID)
	if !ok {
		return
	}

	for i, constraint := range s.constraints {
		cur := s.indexes[i]
		nodes := []*treemap.Map{cur}
		keys := []primitive.Value{nil}

		for i, k := range constraint.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(*treemap.Map)
			if !ok {
				nodes = nil
				keys = nil

				break
			}

			nodes = append(nodes, child)
			keys = append(keys, value)

			if i < len(constraint.Keys)-1 {
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

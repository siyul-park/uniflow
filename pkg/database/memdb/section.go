package memdb

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/tidwall/btree"
)

type Section struct {
	data        *btree.BTreeG[node]
	indexes     []*btree.BTreeG[index]
	constraints []Constraint
	mu          sync.RWMutex
}

type Constraint struct {
	Name    string
	Keys    []string
	Unique  bool
	Partial func(object.Map) bool
}

type node struct {
	key   object.Object
	value object.Map
}

type index struct {
	key   object.Object
	value *btree.BTreeG[index]
}

var (
	nodePool = sync.Pool{
		New: func() any {
			return btree.NewBTreeG[node](nodeComparator)
		},
	}

	indexPool = sync.Pool{
		New: func() any {
			return btree.NewBTreeG[index](indexComparator)
		},
	}
)

var (
	ErrPKNotFound   = errors.New("primary key is not found")
	ErrPKDuplicated = errors.New("primary key is duplicated")
)

var keyID = object.NewString("id")

var (
	nodeComparator = func(x, y node) bool {
		return object.Compare(x.key, y.key) < 0
	}
	indexComparator = func(x, y index) bool {
		return object.Compare(x.key, y.key) < 0
	}
)

func newSection() *Section {
	s := &Section{
		data: newNodes(),
	}

	primary := Constraint{
		Keys:    []string{"id"},
		Name:    "_id",
		Unique:  true,
		Partial: nil,
	}

	s.constraints = append(s.constraints, primary)
	s.indexes = append(s.indexes, newIndexes())

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
	s.indexes = append(s.indexes, newIndexes())

	var err error
	s.data.Scan(func(n node) bool {
		if err = s.index(n.value); err != nil {
			return false
		}
		return true
	})
	return err
}

func (s *Section) DropConstraint(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, constraint := range s.constraints {
		if constraint.Name == name {
			indexes := s.indexes[i]

			s.constraints = append(s.constraints[:i], s.constraints[i+1:]...)
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)

			deleteIndexes(indexes)
		}
	}

	return nil
}

func (s *Section) Constraints() []Constraint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.constraints[:]
}

func (s *Section) Set(doc object.Map) (object.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := doc.Get(keyID)
	if !ok {
		return nil, errors.WithStack(ErrPKNotFound)
	}

	n := node{key: id, value: doc}

	if _, ok := s.data.Get(n); ok {
		return nil, errors.WithStack(ErrPKDuplicated)
	}

	if err := s.index(doc); err != nil {
		return nil, err
	}
	s.data.Set(n)

	return id, nil
}

func (s *Section) Delete(doc object.Map) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := doc.Get(keyID)
	if !ok {
		return false
	}

	n := node{key: id}

	if _, ok := s.data.Get(n); !ok {
		return false
	}

	s.unindex(doc)
	s.data.Delete(n)

	return true
}

func (s *Section) Range(f func(doc object.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.data.Scan(func(n node) bool {
		return f(n.value)
	})
}

func (s *Section) Scan(name string, min, max object.Object) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i, constraint := range s.constraints {
		if constraint.Name != name {
			continue
		}

		sector := &Sector{
			keys:  constraint.Keys,
			data:  s.data,
			index: s.indexes[i],
			mu:    &s.mu,
		}
		return sector.scan(min, max)
	}

	return nil, false
}

func (s *Section) Drop() []object.Map {
	s.mu.Lock()
	defer s.mu.Unlock()

	var data []object.Map
	s.data.Scan(func(n node) bool {
		data = append(data, n.value)
		return true
	})

	s.data.Clear()
	for _, i := range s.indexes {
		i.Clear()
	}

	return data
}

func (s *Section) index(doc object.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrPKNotFound)
	}

	for i, constraint := range s.constraints {
		partial := constraint.Partial
		if partial != nil && !partial(doc) {
			continue
		}

		cur := s.indexes[i]

		for i, k := range constraint.Keys {
			value, _ := object.Pick[object.Object](doc, k)

			child, ok := cur.Get(index{key: value})
			if !ok {
				child = index{key: value, value: newIndexes()}
				cur.Set(child)
			}

			if i < len(constraint.Keys)-1 {
				cur = child.value
			} else {
				child.value.Set(index{key: id})
				if constraint.Unique && child.value.Len() > 1 {
					s.unindex(doc)
					return errors.WithStack(ErrIndexConflict)
				}
			}
		}
	}

	return nil
}

func (s *Section) unindex(doc object.Map) {
	id, ok := doc.Get(keyID)
	if !ok {
		return
	}

	for i, constraint := range s.constraints {
		cur := s.indexes[i]

		paths := []index{{value: cur}}

		for i, k := range constraint.Keys {
			value, _ := object.Pick[object.Object](doc, k)

			child, ok := cur.Get(index{key: value})
			if !ok {
				paths = nil
				break
			}

			if i < len(constraint.Keys)-1 {
				cur = child.value
			} else {
				child.value.Delete(index{key: id})
			}

			paths = append(paths, child)
		}

		for i := len(paths) - 1; i >= 0; i-- {
			child := paths[i]

			if child.value.Len() == 0 && i > 0 {
				parent := paths[i-1]
				parent.value.Delete(child)

				deleteIndexes(child.value)
			}
		}
	}
}

func newNodes() *btree.BTreeG[node] {
	return nodePool.Get().(*btree.BTreeG[node])
}

func deleteNodes(v *btree.BTreeG[node]) {
	v.Clear()
	nodePool.Put(v)
}

func newIndexes() *btree.BTreeG[index] {
	return indexPool.Get().(*btree.BTreeG[index])
}

func deleteIndexes(v *btree.BTreeG[index]) {
	v.Clear()
	indexPool.Put(v)
}

package store

import (
	"sync"

	"github.com/google/btree"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
)

type scanner interface {
	Scan(key types.String, min, max types.Value) scanner
	Range() func(func(types.Value, types.Map) bool)
}

type section struct {
	entries *btree.BTreeG[*entry]
	indexes []*index
	mu      sync.RWMutex
}

type sector struct {
	entries *btree.BTreeG[*entry]
	indexes []*index
	mu      *sync.RWMutex
}

type index struct {
	Keys   []types.String
	Unique bool
	Filter func(types.Map) bool
	nodes  *btree.BTreeG[*node]
}

type entry struct {
	key   types.Value
	value types.Map
}

type node struct {
	key   types.Value
	value *btree.BTreeG[*node]
}

func (e *entry) Less(than btree.Item) bool {
	return types.Compare(e.key, than.(*entry).key) < 0
}

func (n *node) Less(than btree.Item) bool {
	return types.Compare(n.key, than.(*node).key) < 0
}

func newSection() *section {
	s := &section{
		entries: btree.NewG[*entry](2, func(x, y *entry) bool {
			return types.Compare(x.key, y.key) < 0
		}),
	}
	_ = s.Index(&index{Keys: []types.String{types.NewString("id")}, Unique: true})
	return s
}

func (s *section) Index(idx *index) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.indexes); i++ {
		if s.indexes[i] == idx {
			return nil
		}
	}

	idx.nodes = btree.NewG[*node](2, func(x, y *node) bool {
		return types.Compare(x.key, y.key) < 0
	})
	s.indexes = append(s.indexes, idx)

	var err error
	s.entries.Ascend(func(e *entry) bool {
		err = s.index(idx, e.value)
		return err == nil
	})
	return nil
}

func (s *section) Unindex(idx *index) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.indexes); i++ {
		if s.indexes[i] == idx {
			s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *section) Indexes() []*index {
	s.mu.RLock()
	defer s.mu.RUnlock()

	indexes := make([]*index, 0, len(s.indexes))
	for i := 0; i < len(s.indexes); i++ {
		indexes = append(indexes, s.indexes[i])
	}
	return indexes
}

func (s *section) Store(doc types.Map) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessage(ErrKeyMissing, "key: id")
	}

	if s.entries.Has(&entry{key: id}) {
		return errors.WithMessagef(ErrKeyDuplicate, "key: %v", id.Interface())
	}

	s.entries.ReplaceOrInsert(&entry{key: id, value: doc})

	for _, idx := range s.indexes {
		if err := s.index(idx, doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *section) Swap(doc types.Map) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessage(ErrKeyMissing, "key: id")
	}

	old, ok := s.entries.Get(&entry{key: id})
	if !ok {
		return errors.WithMessagef(ErrKeyNotFound, "key: %v", id.Interface())
	}

	s.entries.ReplaceOrInsert(&entry{key: id, value: doc})

	for _, idx := range s.indexes {
		if err := s.unindex(idx, old.value); err != nil {
			return err
		}
		if err := s.index(idx, doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *section) Delete(id types.Value) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.entries.Delete(&entry{key: id})
	if !ok {
		return errors.WithMessagef(ErrKeyNotFound, "key: %v", id.Interface())
	}

	for _, idx := range s.indexes {
		if err := s.unindex(idx, l.value); err != nil {
			return err
		}
	}

	return nil
}

func (s *section) Load(id types.Value) (types.Map, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l, ok := s.entries.Get(&entry{key: id})
	if !ok {
		return nil, errors.WithMessagef(ErrKeyNotFound, "key: %v", id)
	}
	return l.value, nil
}

func (s *section) Scan(key types.String, min, max types.Value) scanner {
	c := &sector{
		entries: s.entries,
		indexes: s.indexes,
		mu:      &s.mu,
	}
	return c.Scan(key, min, max)
}

func (s *section) Range() func(func(types.Value, types.Map) bool) {
	return func(yield func(key types.Value, doc types.Map) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		s.entries.Ascend(func(e *entry) bool {
			return yield(e.key, e.value)
		})
	}
}

func (s *section) index(idx *index, doc types.Map) error {
	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessage(ErrKeyMissing, "key: id")
	}

	if idx.Filter != nil && !idx.Filter(doc) {
		return nil
	}

	curr := idx.nodes
	for i, key := range idx.Keys {
		val := doc.Get(key)

		next, ok := curr.Get(&node{key: val})
		if !ok {
			next = &node{
				key: val,
				value: btree.NewG[*node](2, func(x, y *node) bool {
					return types.Compare(x.key, y.key) < 0
				}),
			}
			curr.ReplaceOrInsert(next)
		}

		if i == len(idx.Keys)-1 {
			if idx.Unique && next.value.Len() > 0 {
				return errors.WithMessagef(ErrKeyDuplicate, "key: %v", val.Interface())
			}
			next.value.ReplaceOrInsert(&node{key: id})
			continue
		}
		curr = next.value
	}
	return nil
}

func (s *section) unindex(idx *index, doc types.Map) error {
	id := doc.Get(types.NewString("id"))
	if id == nil {
		return errors.WithMessage(ErrKeyMissing, "key: id")
	}

	curr := idx.nodes
	nodes := []*node{{value: curr}}
	for i, key := range idx.Keys {
		val := doc.Get(key)

		next, ok := curr.Get(&node{key: val})
		if !ok {
			break
		}

		if i == len(idx.Keys)-1 {
			next.value.Delete(&node{key: id})
		}

		curr = next.value
		nodes = append(nodes, next)
	}

	for i := len(nodes) - 1; i >= 1; i-- {
		curr := nodes[i]
		if curr.value.Len() == 0 {
			parent := nodes[i-1]
			parent.value.Delete(curr)
		}
	}

	return nil
}

func (s *sector) Scan(key types.String, min, max types.Value) scanner {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var indexes []*index
	for _, idx := range s.indexes {
		if len(idx.Keys) == 0 || idx.Keys[0] != key {
			continue
		}

		if max != nil {
			idx.nodes.DescendLessOrEqual(&node{key: max}, func(n *node) bool {
				if min != nil && types.Compare(n.key, min) < 0 {
					return false
				}
				indexes = append(indexes, &index{
					Keys:  idx.Keys[1:],
					nodes: n.value,
				})
				return true
			})
		} else {
			idx.nodes.AscendGreaterOrEqual(&node{key: min}, func(n *node) bool {
				if max != nil && types.Compare(n.key, max) > 0 {
					return false
				}
				indexes = append(indexes, &index{
					Keys:  idx.Keys[1:],
					nodes: n.value,
				})
				return true
			})
		}
	}

	return &sector{
		entries: s.entries,
		indexes: indexes,
		mu:      s.mu,
	}
}

func (s *sector) Range() func(func(types.Value, types.Map) bool) {
	return func(yield func(key types.Value, doc types.Map) bool) {
		s.mu.RLock()

		var indexes []*index
		curr := s.indexes
		for {
			var next []*index
			for _, idx := range curr {
				if len(idx.Keys) == 0 {
					indexes = append(indexes, idx)
					continue
				}
				idx.nodes.Ascend(func(n *node) bool {
					next = append(next, &index{
						Keys:  idx.Keys[1:],
						nodes: n.value,
					})
					return true
				})
			}
			if len(next) == 0 {
				break
			}
			curr = next
		}

		entries := btree.NewG[*entry](2, func(x, y *entry) bool {
			return types.Compare(x.key, y.key) < 0
		})
		for _, idx := range indexes {
			idx.nodes.Ascend(func(n *node) bool {
				e, _ := s.entries.Get(&entry{key: n.key})
				entries.ReplaceOrInsert(e)
				return true
			})
		}

		s.mu.RUnlock()

		entries.Ascend(func(e *entry) bool {
			return yield(e.key, e.value)
		})
	}
}

package store

import (
	"github.com/google/btree"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
	"sync"
)

type Section struct {
	entries *btree.BTreeG[*entry]
	indexes []*index
	mu      sync.RWMutex
}

type Cursor interface {
	Scan(key types.String, min, max types.Value) Cursor
	Range() func(func(types.Value, types.Map) bool)
}

type cursor struct {
	entries *btree.BTreeG[*entry]
	indexes []*index
	mu      *sync.RWMutex
}

type index struct {
	keys   []types.String
	nodes  *btree.BTreeG[*node]
	unique bool
	filter func(types.Map) bool
}

type entry struct {
	key   types.Value
	value types.Map
}

type node struct {
	key   types.Value
	value *btree.BTreeG[*node]
}

var primaryKey = types.NewString("id")

var (
	ErrKeyMissing   = errors.New("key is missing")
	ErrKeyDuplicate = errors.New("key already exists")
	ErrKeyNotFound  = errors.New("key not found")
)

func (e *entry) Less(than btree.Item) bool {
	return types.Compare(e.key, than.(*entry).key) < 0
}

func (n *node) Less(than btree.Item) bool {
	return types.Compare(n.key, than.(*node).key) < 0
}

func WithUnique(unique bool) func(*index) {
	return func(idx *index) {
		idx.unique = unique
	}
}

func WithFilter(filter func(types.Map) bool) func(*index) {
	return func(idx *index) {
		idx.filter = filter
	}
}

func NewSection() *Section {
	s := &Section{
		entries: btree.NewG[*entry](2, func(x, y *entry) bool {
			return types.Compare(x.key, y.key) < 0
		}),
	}
	_ = s.Index([]types.String{primaryKey}, WithUnique(true))
	return s
}

func (s *Section) Index(keys []types.String, options ...func(*index)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := &index{
		keys: keys,
		nodes: btree.NewG[*node](2, func(x, y *node) bool {
			return types.Compare(x.key, y.key) < 0
		}),
	}

	for _, opt := range options {
		opt(idx)
	}

	s.indexes = append(s.indexes, idx)

	var err error
	s.entries.Ascend(func(e *entry) bool {
		err = s.index(idx, e.value)
		return err == nil
	})
	return nil
}

func (s *Section) Unindex(keys []types.String) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.indexes); i++ {
		idx := s.indexes[i]

		if len(keys) != len(idx.keys) {
			continue
		}
		for j := 0; j < len(keys); j++ {
			if !types.Equal(keys[j], idx.keys[j]) {
				continue
			}
		}

		var err error
		s.entries.Ascend(func(e *entry) bool {
			err = s.unindex(idx, e.value)
			return err == nil
		})

		s.indexes = append(s.indexes[:i], s.indexes[i+1:]...)
		break
	}
	return nil
}

func (s *Section) Store(doc types.Map) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pk := doc.Get(primaryKey)
	if pk == nil {
		return errors.WithMessagef(ErrKeyMissing, "key: %s", primaryKey.String())
	}

	if s.entries.Has(&entry{key: pk}) {
		return errors.WithMessagef(ErrKeyDuplicate, "key: %v", pk.Interface())
	}

	s.entries.ReplaceOrInsert(&entry{key: pk, value: doc})

	for _, idx := range s.indexes {
		if err := s.index(idx, doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Section) Delete(pk types.Value) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.entries.Delete(&entry{key: pk})
	if !ok {
		return errors.WithMessagef(ErrKeyNotFound, "key: %v", pk)
	}

	for _, idx := range s.indexes {
		if err := s.unindex(idx, l.value); err != nil {
			return err
		}
	}

	return nil
}

func (s *Section) Load(pk types.Value) (types.Map, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l, ok := s.entries.Get(&entry{key: pk})
	if !ok {
		return nil, errors.WithMessagef(ErrKeyNotFound, "key: %v", pk)
	}
	return l.value, nil
}

func (s *Section) Scan(key types.String, min, max types.Value) Cursor {
	c := &cursor{
		entries: s.entries,
		indexes: s.indexes,
		mu:      &s.mu,
	}
	return c.Scan(key, min, max)
}

func (s *Section) Range() func(func(types.Value, types.Map) bool) {
	return func(yield func(key types.Value, doc types.Map) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		s.entries.Ascend(func(e *entry) bool {
			return yield(e.key, e.value)
		})
	}
}

func (s *Section) index(idx *index, doc types.Map) error {
	pk := doc.Get(primaryKey)
	if pk == nil {
		return errors.WithMessagef(ErrKeyMissing, "key: %s", primaryKey.String())
	}

	if idx.filter != nil && !idx.filter(doc) {
		return nil
	}

	curr := idx.nodes
	for i, key := range idx.keys {
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

		if i == len(idx.keys)-1 {
			if idx.unique && next.value.Len() > 0 {
				return errors.WithMessagef(ErrKeyDuplicate, "key: %v", val.Interface())
			}
			next.value.ReplaceOrInsert(&node{key: pk})
			continue
		}
		curr = next.value
	}
	return nil
}

func (s *Section) unindex(idx *index, doc types.Map) error {
	pk := doc.Get(primaryKey)
	if pk == nil {
		return errors.WithMessagef(ErrKeyMissing, "key: %s", primaryKey.String())
	}

	curr := idx.nodes
	nodes := []*node{{value: curr}}
	for i, key := range idx.keys {
		val := doc.Get(key)

		next, ok := curr.Get(&node{key: val})
		if !ok {
			break
		}

		if i == len(idx.keys)-1 {
			next.value.Delete(&node{key: pk})
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

func (c *cursor) Scan(key types.String, min, max types.Value) Cursor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var indexes []*index
	for _, idx := range c.indexes {
		if len(idx.keys) == 0 || idx.keys[0] != key {
			continue
		}

		idx.nodes.AscendGreaterOrEqual(&node{key: min}, func(n *node) bool {
			if max != nil && types.Compare(n.key, max) > 0 {
				return false
			}
			indexes = append(indexes, &index{
				keys:  idx.keys[1:],
				nodes: n.value,
			})
			return true
		})
	}

	return &cursor{
		entries: c.entries,
		indexes: indexes,
		mu:      c.mu,
	}
}

func (c *cursor) Range() func(func(types.Value, types.Map) bool) {
	return func(yield func(key types.Value, doc types.Map) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		var indexes []*index

		curr := c.indexes
		for {
			var next []*index
			for _, idx := range curr {
				if len(idx.keys) == 0 {
					indexes = append(indexes, idx)
					continue
				}
				idx.nodes.Ascend(func(n *node) bool {
					next = append(next, &index{
						keys:  idx.keys[1:],
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
				e, _ := c.entries.Get(&entry{key: n.key})
				entries.ReplaceOrInsert(e)
				return true
			})
		}

		entries.Ascend(func(e *entry) bool {
			return yield(e.key, e.value)
		})
	}
}

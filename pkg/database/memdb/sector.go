package memdb

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/tidwall/btree"
)

type Sector struct {
	keys  []string
	data  *btree.BTreeG[node]
	index *btree.BTreeG[index]
	mu    *sync.RWMutex
}

func (s *Sector) Range(f func(doc *object.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.scan(nil, nil)
	}

	s.index.Scan(func(i index) bool {
		if n, ok := s.data.Get(node{key: i.key}); !ok {
			return true
		} else {
			return f(n.value)
		}
	})
}

func (s *Sector) Scan(key string, min, max object.Object) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.keys) == 0 || s.keys[0] != key {
		return nil, false
	}

	return s.scan(min, max)
}

func (s *Sector) scan(min, max object.Object) (*Sector, bool) {
	child := newIndexes()

	if min != nil && max != nil && object.Compare(min, max) == 0 {
		if i, ok := s.index.Get(index{key: min}); ok {
			child = i.value
		}
	} else {
		s.index.Ascend(index{key: min}, func(i index) bool {
			k := i.key
			v := i.value

			if max != nil && object.Compare(k, max) > 0 {
				return false
			}

			v.Scan(func(c index) bool {
				if l, ok := child.Get(c); ok {
					c.value = deepMerge(c.value, l.value)
				}
				child.Set(c)
				return true
			})
			return true
		})
	}

	return &Sector{
		data:  s.data,
		keys:  s.keys[1:],
		index: child,
		mu:    s.mu,
	}, true
}

func deepMerge(x, y *btree.BTreeG[index]) *btree.BTreeG[index] {
	merged := newIndexes()

	x.Scan(func(i index) bool {
		merged.Set(i)
		return true
	})
	y.Scan(func(i index) bool {
		if l, ok := merged.Get(i); ok {
			i.value = deepMerge(i.value, l.value)
		}
		merged.Set(i)
		return true
	})

	return merged
}

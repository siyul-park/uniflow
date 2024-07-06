package memdb

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/tidwall/btree"
)

// Sector represents a sector within a Section, facilitating range scans and traversal.
type Sector struct {
	keys  []string
	data  *btree.BTreeG[node]
	index *btree.BTreeG[index]
	mu    *sync.RWMutex
}

// Range iterates over all documents in the sector and applies the given function.
func (s *Sector) Range(f func(doc types.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.scan(nil, nil)
	}

	s.index.Scan(func(i index) bool {
		if n, ok := s.data.Get(node{key: i.key}); ok {
			return f(n.value)
		}
		return true
	})
}

// Scan performs a range scan on the sector using the specified key, min, and max values.
// It returns a new sector and a boolean indicating if the scan was successful.
func (s *Sector) Scan(key string, min, max types.Object) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.keys) == 0 || s.keys[0] != key {
		return nil, false
	}

	return s.scan(min, max)
}

func (s *Sector) scan(min, max types.Object) (*Sector, bool) {
	child := newIndexes()

	if min != nil && max != nil && types.Compare(min, max) == 0 {
		if i, ok := s.index.Get(index{key: min}); ok {
			child = i.value
		}
	} else {
		s.index.Ascend(index{key: min}, func(i index) bool {
			k := i.key
			v := i.value

			if max != nil && types.Compare(k, max) > 0 {
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

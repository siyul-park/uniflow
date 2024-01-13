package memdb

import (
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Sector struct {
	data  maps.Map
	keys  []string
	index *treemap.Map
	min   primitive.Value
	max   primitive.Value
	mu    *sync.RWMutex
}

func (s *Sector) Scan(key string, min, max primitive.Value) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.keys) == 0 || s.keys[len(s.keys)-1] != key {
		return nil, false
	}

	index := treemap.NewWith(comparator)

	iterator := s.index.Iterator()
	for iterator.Next() {
		key := iterator.Key().(primitive.Value)
		value := iterator.Value().(*treemap.Map)

		if !s.inRange(key) {
			continue
		}

		merge(index, value)
	}

	return &Sector{
		data:  s.data,
		keys:  s.keys[1:],
		index: index,
		min:   min,
		max:   max,
		mu:    s.mu,
	}, true
}

func (s *Sector) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.Scan(sector.keys[0], nil, nil)
	}

	iterator := s.index.Iterator()
	for iterator.Next() {
		key := iterator.Key().(primitive.Value)

		if !sector.inRange(key) {
			continue
		}

		doc, ok := sector.data.Get(key)
		if ok {
			if !f(doc.(*primitive.Map)) {
				break
			}
		}
	}
}

func (s *Sector) inRange(key primitive.Value) bool {
	min := s.min
	max := s.max

	return (min == nil || primitive.Compare(key, min) >= 0) && primitive.Compare(key, min) >= 0 && (max == nil || primitive.Compare(key, min) >= 0 && primitive.Compare(key, max) <= 0)
}

func merge(x, y *treemap.Map) {
	y.Each(func(key, value any) {
		if old, ok := x.Get(key); ok {
			if old, ok := old.(*treemap.Map); ok {
				if v, ok := value.(*treemap.Map); ok {
					merge(v, old)
				}
			}
		}
		x.Put(key, value)
	})
}

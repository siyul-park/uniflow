package memdb

import (
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Sector struct {
	keys  []string
	data  maps.Map
	index *treemap.Map
	mu    *sync.RWMutex
}

func newSector(
	keys []string,
	data maps.Map,
	index *treemap.Map,
	min primitive.Value,
	max primitive.Value,
	mu *sync.RWMutex,
) *Sector {
	mu.RLock()
	defer mu.RUnlock()

	if len(keys) == 0 {
		return &Sector{
			data:  data,
			index: index,
			mu:    mu,
		}
	}

	if min != nil && max != nil && primitive.Compare(min, max) == 0 {
		value, ok := index.Get(min)
		if !ok {
			value = treemap.NewWith(comparator)
		}

		return &Sector{
			data:  data,
			keys:  keys[1:],
			index: value.(*treemap.Map),
			mu:    mu,
		}
	}

	child := treemap.NewWith(comparator)

	for iterator := index.Iterator(); iterator.Next(); {
		key := iterator.Key().(primitive.Value)
		value := iterator.Value().(*treemap.Map)

		if (min != nil && primitive.Compare(key, min) < 0) || (max != nil && primitive.Compare(key, max) > 0) {
			continue
		}

		value.Each(func(key, value any) {
			if v, ok := value.(*treemap.Map); ok {
				if old, ok := child.Get(key); ok {
					value = mergeMap(old.(*treemap.Map), v)
				}
			}
			child.Put(key, value)
		})
	}

	return &Sector{
		data:  data,
		keys:  keys[1:],
		index: child,
		mu:    mu,
	}
}

func (s *Sector) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.Scan(sector.keys[0], nil, nil)
	}

	for iterator := sector.index.Iterator(); iterator.Next(); {
		key, _ := iterator.Key().(primitive.Value)

		doc, ok := s.data.Get(key)
		if ok {
			if !f(doc.(*primitive.Map)) {
				return
			}
		}
	}
}

func (s *Sector) Scan(key string, min, max primitive.Value) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.keys) == 0 || s.keys[0] != key {
		return nil, false
	}
	return newSector(s.keys, s.data, s.index, min, max, s.mu), true
}

func mergeMap(x, y *treemap.Map) *treemap.Map {
	z := treemap.NewWith(comparator)

	x.Each(func(key, value any) {
		z.Put(key, value)
	})
	y.Each(func(key, value any) {
		if old, ok := z.Get(key); ok {
			if old, ok := old.(*treemap.Map); ok {
				if v, ok := value.(*treemap.Map); ok {
					value = mergeMap(old, v)
				}
			}
		}
		z.Put(key, value)
	})

	return z
}

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

func (s *Sector) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.Scan(sector.keys[0], nil, nil)
	}

	for iterator := sector.index.Iterator(); iterator.Next(); {
		key := iterator.Key()
		doc, ok := s.data.Get(key)
		if ok {
			if !f(doc.(*primitive.Map)) {
				break
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

	if min != nil && max != nil && primitive.Compare(min, max) == 0 {
		value, ok := s.index.Get(min)
		if !ok {
			value = treemap.NewWith(comparator)
		}

		return &Sector{
			data:  s.data,
			keys:  s.keys[1:],
			index: value.(*treemap.Map),
			mu:    s.mu,
		}, true
	}

	index := treemap.NewWith(comparator)

	s.index.Each(func(key, value any) {
		k := key.(primitive.Value)
		v := value.(*treemap.Map)

		if (min != nil && primitive.Compare(k, min) < 0) || (max != nil && primitive.Compare(k, max) > 0) {
			return
		}

		v.Each(func(key, value any) {
			if v, ok := value.(*treemap.Map); ok {
				if old, ok := index.Get(key); ok {
					value = mergeMap(old.(*treemap.Map), v)
				}
			}
			index.Put(key, value)
		})
	})

	return &Sector{
		data:  s.data,
		keys:  s.keys[1:],
		index: index,
		mu:    s.mu,
	}, true
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

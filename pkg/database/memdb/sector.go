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

func (s *Sector) Range(f func(doc *primitive.Map) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sector := s
	for len(sector.keys) > 0 {
		sector, _ = sector.Scan(sector.keys[0], nil, nil)
	}

	if v := sector.single(); v != nil {
		value, ok := sector.index.Get(v)
		if !ok {
			return
		}
		sector.each(value.(*treemap.Map), f)
		return
	}

	for iterator := sector.index.Iterator(); iterator.Next(); {
		key := iterator.Key().(primitive.Value)
		value := iterator.Value().(*treemap.Map)
		if !sector.inRange(key) {
			continue
		}
		sector.each(value, f)
	}
}

func (s *Sector) Scan(key string, min, max primitive.Value) (*Sector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.keys) == 0 || s.keys[0] != key {
		return nil, false
	}

	if v := s.single(); v != nil {
		value, ok := s.index.Get(v)
		if !ok {
			value = treemap.NewWith(comparator)
		}
		return &Sector{
			data:  s.data,
			keys:  s.keys[1:],
			index: value.(*treemap.Map),
			min:   min,
			max:   max,
			mu:    s.mu,
		}, true
	}

	index := treemap.NewWith(comparator)

	for iterator := s.index.Iterator(); iterator.Next(); {
		key := iterator.Key().(primitive.Value)
		value := iterator.Value().(*treemap.Map)
		if !s.inRange(key) {
			continue
		}
		value.Each(func(key, value any) {
			v, _ := value.(*treemap.Map)
			if old, ok := index.Get(key); ok {
				v = s.mergeMap(old.(*treemap.Map), v)
			}
			index.Put(key, v)
		})
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

func (s *Sector) each(value *treemap.Map, f func(doc *primitive.Map) bool) {
	for iterator := value.Iterator(); iterator.Next(); {
		key := iterator.Key().(primitive.Value)
		doc, ok := s.data.Get(key)
		if ok {
			if !f(doc.(*primitive.Map)) {
				return
			}
		}
	}
}

func (s *Sector) single() primitive.Value {
	if s.min != nil && s.max != nil && primitive.Compare(s.min, s.max) == 0 {
		return s.min
	}
	return nil
}

func (s *Sector) inRange(key primitive.Value) bool {
	min := s.min
	max := s.max

	return (min == nil || primitive.Compare(key, min) >= 0) && (max == nil || primitive.Compare(key, max) <= 0)
}

func (s *Sector) mergeMap(x, y *treemap.Map) *treemap.Map {
	z := treemap.NewWith(comparator)

	x.Each(func(key, value any) {
		z.Put(key, value)
	})
	y.Each(func(key, value any) {
		if old, ok := z.Get(key); ok {
			if old, ok := old.(*treemap.Map); ok {
				if v, ok := value.(*treemap.Map); ok {
					value = s.mergeMap(old, v)
				}
			}
		}
		z.Put(key, value)
	})

	return z
}

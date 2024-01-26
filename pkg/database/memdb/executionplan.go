package memdb

import (
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type executionPlan struct {
	key  string
	min  primitive.Value
	max  primitive.Value
	next *executionPlan
}

func newExecutionPlan(keys []string, filter *database.Filter) *executionPlan {
	if filter == nil {
		return nil
	}

	var plan *executionPlan

	switch filter.OP {
	case database.AND:
		for _, child := range filter.Children {
			plan = plan.intersect(newExecutionPlan(keys, child))
		}
	case database.OR:
		for _, child := range filter.Children {
			plan = plan.union(newExecutionPlan(keys, child))
		}
	case database.IN:
		value := filter.Value.(*primitive.Slice)
		for _, v := range value.Values() {
			plan = plan.union(newExecutionPlan(keys, database.Where(filter.Key).EQ(v)))
		}
	case database.EQ, database.GT, database.GTE, database.LT, database.LTE:
		var pre *executionPlan
		for _, key := range keys {
			var cur *executionPlan
			if key != filter.Key {
				cur = &executionPlan{
					key: key,
				}
			} else {
				value := filter.Value

				var min primitive.Value
				var max primitive.Value
				if filter.OP == database.EQ {
					min = value
					max = value
				} else if filter.OP == database.GT || filter.OP == database.GTE {
					min = value
				} else if filter.OP == database.LT || filter.OP == database.LTE {
					max = value
				}

				cur = &executionPlan{
					key: key,
					min: min,
					max: max,
				}
			}

			if pre == nil {
				plan = cur
			} else {
				pre.next = cur
			}
			pre = cur

			if cur.min != nil || cur.max != nil {
				break
			}
		}
		if pre != nil && pre.min == nil && pre.max == nil {
			plan = nil
		}
	}

	return plan
}

func (e *executionPlan) Cost() int {
	cur := 0
	if e.min == nil || e.max == nil || primitive.Compare(e.min, e.max) != 0 {
		cur = 1
	}

	if e.next != nil {
		cur += e.next.Cost()
	}

	return cur
}

func (e *executionPlan) intersect(other *executionPlan) *executionPlan {
	if e == nil {
		return other
	}
	if other == nil {
		return e
	}

	if e.key != other.key {
		return nil
	}

	z := &executionPlan{
		key: e.key,
	}

	if e.min == nil {
		z.min = other.min
	} else if primitive.Compare(e.min, other.min) < 0 {
		z.min = other.min
	} else {
		z.min = e.min
	}

	if e.max == nil {
		z.max = nil
	} else if primitive.Compare(e.max, other.max) > 0 {
		z.max = other.max
	} else {
		z.max = e.max
	}

	z.next = e.next.intersect(other.next)

	return z
}

func (e *executionPlan) union(other *executionPlan) *executionPlan {
	if e == nil || other == nil || e.key != other.key {
		return nil
	}

	z := &executionPlan{
		key: e.key,
	}

	if e.min == nil || z.min == nil {
		z.min = nil
	} else if primitive.Compare(e.min, other.min) > 0 {
		z.min = other.min
	} else {
		z.min = e.min
	}

	if e.max == nil || z.max == nil {
		z.max = nil
	} else if primitive.Compare(e.max, other.max) < 0 {
		z.max = other.max
	} else {
		z.max = e.max
	}

	z.next = e.next.union(other.next)

	allNil := true
	for cur := z; cur != nil; cur = cur.next {
		if cur.min != nil || cur.max != nil {
			allNil = false
			break
		}
	}
	if allNil {
		return nil
	}

	return z
}

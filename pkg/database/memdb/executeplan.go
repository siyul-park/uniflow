package memdb

import (
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type executePlan struct {
	key  string
	min  primitive.Value
	max  primitive.Value
	next *executePlan
}

func buildExecutePlan(keys []string, filter *database.Filter) *executePlan {
	if filter == nil {
		return nil
	}

	var plan *executePlan

	switch filter.OP {
	case database.AND:
		for _, child := range filter.Children {
			plan = plan.and(buildExecutePlan(keys, child))
		}
	case database.OR:
		for _, child := range filter.Children {
			plan = plan.or(buildExecutePlan(keys, child))
		}
	case database.IN:
		value := filter.Value.(*primitive.Slice)
		for _, v := range value.Values() {
			plan = plan.or(buildExecutePlan(keys, database.Where(filter.Key).EQ(v)))
		}
	case database.EQ, database.GT, database.GTE, database.LT, database.LTE:
		var root *executePlan
		var pre *executePlan

		for _, key := range keys {
			if key != filter.Key {
				p := &executePlan{
					key: key,
				}

				if pre != nil {
					pre.next = p
				} else {
					root = p
				}
				pre = p
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

				p := &executePlan{
					key: key,
					min: min,
					max: max,
				}

				if pre != nil {
					pre.next = p
				} else {
					root = p
				}
				plan = root
				break
			}
		}
	}

	return plan
}

func (e *executePlan) and(other *executePlan) *executePlan {
	if e == nil {
		return other
	}
	if other == nil {
		return e
	}

	if e.key != other.key {
		return nil
	}

	z := &executePlan{
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

	z.next = e.next.and(other.next)
	if e.next != other.next && z.next == nil {
		return nil
	}

	return z
}

func (e *executePlan) or(other *executePlan) *executePlan {
	if e == nil || other == nil || e.key != other.key {
		return nil
	}

	z := &executePlan{
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

	z.next = e.next.or(other.next)

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

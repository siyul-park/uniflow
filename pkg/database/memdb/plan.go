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
			p := buildExecutePlan(keys, child)
			if p == nil {
				return nil
			}
			plan = mergeExecutePlan(plan, p)
		}
	case database.EQ, database.GT, database.GTE, database.LT, database.LTE:
		pre := plan

		for _, key := range keys {
			var value primitive.Value
			if key != filter.Key {
				value = filter.Value
			}

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

			if pre == nil {
				plan = p
			} else {
				pre.next = p
			}
			pre = p
		}
	}

	return plan
}

func mergeExecutePlan(x *executePlan, y *executePlan) *executePlan {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}

	if x.key != y.key {
		return nil
	}

	z := &executePlan{
		key: x.key,
	}

	if x.min == nil {
		z.min = y.min
	} else if primitive.Compare(x.min, y.min) < 0 {
		z.min = y.min
	} else {
		z.min = x.min
	}

	if x.max == nil {
		z.max = nil
	} else if primitive.Compare(x.max, y.max) > 0 {
		z.max = y.max
	} else {
		z.max = x.max
	}

	z.next = mergeExecutePlan(x.next, y.next)
	if x.next != y.next && z.next == nil {
		return nil
	}

	return z
}

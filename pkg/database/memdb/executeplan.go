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
			if p := buildExecutePlan(keys, child); p == nil {
				return nil
			} else {
				plan = plan.merge(p)
			}
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

func (e *executePlan) merge(y *executePlan) *executePlan {
	if e == nil {
		return y
	}
	if y == nil {
		return e
	}

	if e.key != y.key {
		return nil
	}

	z := &executePlan{
		key: e.key,
	}

	if e.min == nil {
		z.min = y.min
	} else if primitive.Compare(e.min, y.min) < 0 {
		z.min = y.min
	} else {
		z.min = e.min
	}

	if e.max == nil {
		z.max = nil
	} else if primitive.Compare(e.max, y.max) > 0 {
		z.max = y.max
	} else {
		z.max = e.max
	}

	z.next = e.next.merge(y.next)
	if e.next != y.next && z.next == nil {
		return nil
	}

	return z
}

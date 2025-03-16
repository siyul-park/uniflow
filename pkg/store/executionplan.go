package store

import "github.com/siyul-park/uniflow/pkg/types"

type executionPlan struct {
	key  types.String
	min  types.Value
	max  types.Value
	next *executionPlan
}

func newExecutionPlan(keys []types.String, filter types.Value) *executionPlan {
	f, ok := filter.(types.Map)
	if !ok || len(keys) == 0 {
		return nil
	}

	key := keys[0]
	value := f.Get(key)

	plan := &executionPlan{key: key}

	if val, ok := value.(types.Map); ok {
		if v := val.Get(types.NewString("$eq")); v != nil {
			plan.min, plan.max = v, v
		}

		var lowers []types.Value
		if v := val.Get(types.NewString("$gt")); v != nil {
			lowers = append(lowers, v)
		}
		if v := val.Get(types.NewString("$gte")); v != nil {
			lowers = append(lowers, v)
		}
		for _, l := range lowers {
			if plan.min == nil || types.Compare(l, plan.min) > 0 {
				plan.min = l
			}
		}

		var uppers []types.Value
		if v := val.Get(types.NewString("$lt")); v != nil {
			uppers = append(uppers, v)
		}
		if v := val.Get(types.NewString("$lte")); v != nil {
			uppers = append(uppers, v)
		}
		for _, u := range uppers {
			if plan.max == nil || types.Compare(u, plan.max) < 0 {
				plan.max = u
			}
		}
	} else {
		plan.min, plan.max = value, value
	}

	if v, ok := f.Get(types.NewString("$and")).(types.Slice); ok {
		for _, child := range v.Range() {
			plan.intersect(newExecutionPlan(keys, child))
		}
	}
	if v, ok := f.Get(types.NewString("$or")).(types.Slice); ok {
		for _, child := range v.Range() {
			plan.union(newExecutionPlan(keys, child))
		}
	}

	if plan.min == nil && plan.max == nil {
		return nil
	}

	plan.next = newExecutionPlan(keys[1:], filter)
	return plan
}

func (e *executionPlan) intersect(other *executionPlan) {
	if other == nil {
		return
	}
	if e.key != other.key {
		e.min, e.max = nil, nil
		return
	}

	if e.min == nil || types.Compare(other.min, e.min) > 0 {
		e.min = other.min
	}
	if e.max == nil || types.Compare(other.max, e.max) < 0 {
		e.max = other.max
	}
}

func (e *executionPlan) union(other *executionPlan) {
	if other == nil || e.key != other.key {
		e.min, e.max = nil, nil
		return
	}

	if other.min != nil && types.Compare(other.min, e.min) < 0 {
		e.min = other.min
	}
	if other.max != nil && types.Compare(other.max, e.max) > 0 {
		e.max = other.max
	}
}

func (e *executionPlan) lenght() int {
	if e.next != nil {
		return 1 + e.next.lenght()
	}
	return 1
}

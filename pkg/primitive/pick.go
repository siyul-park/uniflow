package primitive

import (
	"strconv"
)

func Pick[T any](v Object, paths ...string) (T, bool) {
	var zero T

	cur := v
	for _, path := range paths {
		switch v := cur.(type) {
		case *Map:
			child, ok := v.Get(NewString(path))
			if !ok {
				return zero, false
			}
			cur = child

		case *Slice:
			index, err := strconv.Atoi(path)
			if err != nil || index >= v.Len() {
				return zero, false
			}
			cur = v.Get(index)
		default:
			return zero, false
		}
	}

	if v, ok := cur.(T); ok {
		return v, true
	} else if cur == nil {
		return zero, false
	} else if v, ok := cur.Interface().(T); ok {
		return v, true
	} else {
		return zero, false
	}
}

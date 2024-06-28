package memdb

import (
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/object"
)

func parseSorts(sorts []database.Sort) func(i, j object.Map) bool {
	return func(i, j object.Map) bool {
		for _, s := range sorts {
			x, _ := object.Pick[object.Object](i, parsePath(s.Key)...)
			y, _ := object.Pick[object.Object](j, parsePath(s.Key)...)

			e := object.Compare(x, y)
			if e == 0 {
				continue
			}

			if s.Order == database.OrderDESC {
				return e > 0
			}
			return e < 0
		}
		return false
	}
}

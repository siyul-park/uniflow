package memdb

import (
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
)

func parseSorts(sorts []database.Sort) func(i, j types.Map) bool {
	return func(i, j types.Map) bool {
		for _, s := range sorts {
			x, _ := types.Pick[types.Value](i, parsePath(s.Key)...)
			y, _ := types.Pick[types.Value](j, parsePath(s.Key)...)

			e := types.Compare(x, y)
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

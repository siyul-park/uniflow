package memdb

import (
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

func ParseSorts(sorts []database.Sort) func(i, j *primitive.Map) bool {
	return func(i, j *primitive.Map) bool {
		for _, s := range sorts {
			x, _ := primitive.Pick[primitive.Object](i, s.Key)
			y, _ := primitive.Pick[primitive.Object](j, s.Key)

			if x == y {
				continue
			} else if x == nil {
				return s.Order == database.OrderDESC
			} else if y == nil {
				return s.Order != database.OrderDESC
			}

			e := util.Compare(x.Interface(), y.Interface())
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

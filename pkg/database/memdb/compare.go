package memdb

import (
	"github.com/emirpasic/gods/utils"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

var (
	comparator = utils.Comparator(func(a, b any) int {
		return primitive.Compare(a.(primitive.Object), b.(primitive.Object))
	})
)

package memdb

import (
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Sector struct {
	keys  []string
	index *treemap.Map
	min   primitive.Value
	max   primitive.Value
	mu    *sync.RWMutex
}

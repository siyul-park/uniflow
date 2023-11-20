package pool

import "sync"

var (
	mapPool = sync.Pool{New: func() any {
		return &sync.Map{}
	}}
)

func GetMap() *sync.Map {
	return mapPool.Get().(*sync.Map)
}

func PutMap(v *sync.Map) {
	v.Range(func(key, _ any) bool {
		v.Delete(key)
		return true
	})
	mapPool.Put(v)
}

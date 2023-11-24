package primitive

type (
	ordered interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
	}
)

func compare[T ordered](x, y T) int {
	if x == y {
		return 0
	}
	if x > y {
		return 1
	}
	if x < y {
		return -1
	}
	return 0
}

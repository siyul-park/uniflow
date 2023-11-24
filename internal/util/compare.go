package util

import (
	"math"
	"reflect"
)

func Compare(x any, y any) int {
	c, ok := compare(reflect.ValueOf(x), reflect.ValueOf(y))
	if !ok {
		return 0
	}
	return c
}

func compare(x, y reflect.Value) (int, bool) {
	x = rawValue(x)
	y = rawValue(y)

	k1 := basicKind(x)
	k2 := basicKind(y)

	if k1 == invalidKind || k2 == invalidKind {
		return 0, false
	}
	if k1 == pointerKind {
		return compare(x.Elem(), y)
	}
	if k2 == pointerKind {
		return compare(x, y.Elem())
	}

	if k1 != k2 {
		switch {
		case k1 == intKind && k2 == uintKind:
			if x.Int() < 0 {
				return -1, true
			}
			return compareStrict(uint64(x.Int()), y.Uint()), true
		case k1 == uintKind && k2 == intKind:
			if y.Int() < 0 {
				return 1, true
			}
			return compareStrict(x.Uint(), uint64(y.Int())), true
		default:
			return compareStrict(k1, k2), true
		}
	} else {
		switch k1 {
		case nullKind:
			return 0, true
		case floatKind:
			return compareStrict(x.Float(), y.Float()), true
		case intKind:
			return compareStrict(x.Int(), y.Int()), true
		case uintKind:
			return compareStrict(x.Uint(), y.Uint()), true
		case stringKind:
			return compareStrict(x.String(), y.String()), true
		case iterableKind:
			for i := 0; i < int(math.Min(float64(x.Len()), float64(y.Len()))); i++ {
				if c, ok := compare(x.Index(i), y.Index(i)); ok && c != 0 {
					return c, true
				} else if !ok {
					return 0, false
				}
			}
			return compareStrict(x.Len(), y.Len()), true
		default:
			return 0, false
		}
	}
}

func compareStrict[T Ordered](x T, y T) int {
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

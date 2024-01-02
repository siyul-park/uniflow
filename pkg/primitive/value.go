package primitive

import "reflect"

// Kind represents the enumeration of data types.
type Kind uint

// Value is an interface that signifies atomic data types.
type Value interface {
	Kind() Kind          // Returns the type of data.
	Compare(v Value) int // Compares with another Value and returns the order.
	Interface() any      // Converts the internal value to a generic interface.
}

// Constants representing various data types.
const (
	KindInvalid Kind = iota
	KindBinary
	KindBool
	KindInt
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat32
	KindFloat64
	KindMap
	KindSlice
	KindString
)

var typeAny = reflect.TypeOf((*any)(nil)).Elem()

// Compare function compares two Values and returns their order.
// Nil values are treated as the lowest order.
func Compare(x, y Value) int {
	if x == nil && y == nil {
		return 0
	} else if x == nil {
		return -1
	} else if y == nil {
		return 1
	} else {
		return x.Compare(y)
	}
}

// Interface function converts a Value to a generic interface.
// Nil values are returned as a nil interface.
func Interface(v Value) any {
	if v == nil {
		return nil
	} else {
		return v.Interface()
	}
}

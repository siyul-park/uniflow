package types

import "reflect"

// Value is an interface representing atomic data types.
type Value interface {
	Kind() Kind              // Kind returns the type of the Value.
	Hash() uint64            // Hash returns the hash code of the Value.
	Interface() any          // Interface returns the Value as a generic interface.
	Equal(other Value) bool  // Equal checks if this Value equals another Value.
	Compare(other Value) int // Compare compares this Value with another Value.
}

// Kind represents enumerated data types.
type Kind byte

// Constants representing various data types.
const (
	KindInvalid Kind = iota
	KindBinary
	KindBoolean
	KindError
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

var types = map[Kind]reflect.Type{
	KindInvalid: reflect.TypeOf((*any)(nil)).Elem(),
	KindBinary:  reflect.TypeOf([]byte(nil)),
	KindBoolean: reflect.TypeOf(false),
	KindError:   reflect.TypeOf((*error)(nil)).Elem(),
	KindInt:     reflect.TypeOf(0),
	KindInt8:    reflect.TypeOf(int8(0)),
	KindInt16:   reflect.TypeOf(int16(0)),
	KindInt32:   reflect.TypeOf(int32(0)),
	KindInt64:   reflect.TypeOf(int64(0)),
	KindUint:    reflect.TypeOf(uint(0)),
	KindUint8:   reflect.TypeOf(uint8(0)),
	KindUint16:  reflect.TypeOf(uint16(0)),
	KindUint32:  reflect.TypeOf(uint32(0)),
	KindUint64:  reflect.TypeOf(uint64(0)),
	KindFloat32: reflect.TypeOf(float32(0)),
	KindFloat64: reflect.TypeOf(float64(0)),
	KindMap:     reflect.TypeOf((*any)(nil)).Elem(),
	KindSlice:   reflect.TypeOf((*any)(nil)).Elem(),
	KindString:  reflect.TypeOf(""),
}

// KindOf returns the kind of the provided Value.
func KindOf(v Value) Kind {
	if v == nil {
		return KindInvalid
	}
	return v.Kind()
}

// TypeOf returns the reflect.Type of the provided Kind.
func TypeOf(kind Kind) reflect.Type {
	return types[kind]
}

// HashOf returns the hash code of the provided Value.
func HashOf(v Value) uint64 {
	if v == nil {
		return 0
	}
	return v.Hash()
}

// InterfaceOf converts a Value to a generic interface.
func InterfaceOf(v Value) any {
	if v == nil {
		return nil
	}
	return v.Interface()
}

// Equal checks equality between two Values.
func Equal(x, y Value) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return x.Equal(y)
}

// Compare compares two Values and returns.
func Compare(x, y Value) int {
	if x == nil && y == nil {
		return 0
	}
	if x == nil {
		return -1
	}
	if y == nil {
		return 1
	}
	return x.Compare(y)
}

func unionType(x, y reflect.Type) reflect.Type {
	if x == nil {
		return y
	} else if y == nil {
		return x
	} else if x == y {
		return x
	}
	return types[KindInvalid]
}

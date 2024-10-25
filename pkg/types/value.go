package types

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

// KindOf returns the kind of the provided Value.
func KindOf(v Value) Kind {
	if v == nil {
		return KindInvalid
	}
	return v.Kind()
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

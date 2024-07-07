package types

// Value is an interface that signifies atomic data types.
type Value interface {
	Kind() Kind              // Kind returns the kind of the Object.
	Hash() uint64            // Hash returns the hash code of the Object.
	Interface() any          // Interface returns the Object as a generic interface.
	Equal(other Value) bool  // Equal checks whether the Object is equal to another Object.
	Compare(other Value) int // Compare compares the Object with another Object.
}

// Kind represents the enumeration of data types.
type Kind byte

// Constants representing various data types.
const (
	KindInvalid Kind = iota // Represents an invalid or nil type.
	KindBinary              // Represents binary data.
	KindBoolean             // Represents a boolean value.
	KindError               // Represents a error value.
	KindInt                 // Represents an integer value.
	KindInt8                // Represents an integer value.
	KindInt16               // Represents an integer value.
	KindInt32               // Represents an integer value.
	KindInt64               // Represents an integer value.
	KindUint                // Represents an unsigned integer value.
	KindUint8               // Represents an unsigned integer value.
	KindUint16              // Represents an unsigned integer value.
	KindUint32              // Represents an unsigned integer value.
	KindUint64              // Represents an unsigned integer value.
	KindFloat32             // Represents a floating-point number.
	KindFloat64             // Represents a floating-point number.
	KindMap                 // Represents a map.
	KindSlice               // Represents a slice.
	KindString              // Represents a string.
)

// KindOf returns the kind of the provided Object.
// If the Object is nil, it returns KindInvalid.
func KindOf(v Value) Kind {
	if v == nil {
		return KindInvalid
	}
	return v.Kind()
}

// HashOf returns the hash code of the provided Object.
// If the Object is nil, it returns 0.
func HashOf(v Value) uint64 {
	if v == nil {
		return 0
	}
	return v.Hash()
}

// InterfaceOf converts an Object to a generic interface.
// Nil values are returned as a nil interface.
func InterfaceOf(v Value) any {
	if v == nil {
		return nil
	}
	return v.Interface()
}

// Equal checks whether two Objects are equal.
// If both Objects are nil, they are considered equal.
// If their kinds differ, they are not equal.
func Equal(x, y Value) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return x.Equal(y)
}

// Compare compares two Objects and returns:
// -1 if x is less than y,
// 0 if they are equal,
// 1 if x is greater than y.
// If both Objects are nil, they are considered equal.
// If their kinds differ, they are not equal.
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

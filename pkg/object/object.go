package object

// Object is an interface that signifies atomic data types.
type Object interface {
	// Kind returns the type of data.
	Kind() Kind

	// Hash returns the hash code for this object.
	// The hash is unique within the same kind.
	Hash() uint64

	// Compare compares this object with another object.
	// Returns an integer indicating the order.
	// Deprecated: Use Hash for comparison instead.
	Compare(v Object) int

	// Interface converts the internal value to a generic interface{}.
	Interface() any
}

// Kind represents the enumeration of data types.
type Kind byte

// Constants representing various data types.
const (
	KindInvalid Kind = iota
	KindBinary
	KindBuffer
	KindBool
	KindInteger
	KindUInteger
	KindFloat
	KindMap
	KindSlice
	KindString
)

// KindOf returns the kind of the provided value.
// If the value is nil, it returns KindInvalid.
// Otherwise, it calls the Kind method of the value to determine its kind.
func KindOf(v Object) Kind {
	if v == nil {
		return KindInvalid
	}
	return v.Kind()
}

// Hash returns the hash code of the provided Object.
// If the Object is nil, it returns 0.
func Hash(v Object) uint64 {
	if v == nil {
		return 0
	}
	return v.Hash()
}

// Compare function compares two Objects and returns their order.
// Nil values are treated as the lowest order.
func Compare(x, y Object) int {
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

// Interface converts an Object to a generic interface.
// Nil values are returned as a nil interface.
func Interface(v Object) any {
	if v == nil {
		return nil
	}
	return v.Interface()
}

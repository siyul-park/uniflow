package primitive

type (
	// Object is an atomic type.
	Object interface {
		Kind() Kind
		Equal(v Object) bool
		Compare(v Object) int
		Hash() uint32
		Interface() any
	}

	Kind uint
)

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

func Equal(x, y Object) bool {
	if x == nil && y == nil {
		return true
	} else if x == nil {
		return false
	} else if y == nil {
		return false
	} else {
		return x.Equal(y)
	}
}

func Compare(x, y Object) int {
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

func Interface(v Object) any {
	if v == nil {
		return nil
	} else {
		return v.Interface()
	}
}

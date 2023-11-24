package primitive

type (
	// Object is an atomic type.
	Object interface {
		Kind() Kind
		Equal(v Object) bool
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


func Interface(v any) any {
	if v == nil {
		return nil
	} else if v, ok := v.(Object); !ok {
		return nil
	} else {
		return v.Interface()
	}
}

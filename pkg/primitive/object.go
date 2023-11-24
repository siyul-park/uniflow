package primitive

import "github.com/siyul-park/uniflow/internal/util"

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

func Interface(v any) any {
	if util.IsNil(v) {
		return nil
	} else if v, ok := v.(Object); !ok {
		return nil
	} else {
		return v.Interface()
	}
}

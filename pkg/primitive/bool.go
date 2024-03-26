package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"reflect"
	"unsafe"
)

// Bool is a representation of a bool.
type Bool bool

var _ Value = (Bool)(false)

var (
	TRUE  = NewBool(true)
	FALSE = NewBool(false)
)

// NewBool returns a new Bool.
func NewBool(value bool) Bool {
	return Bool(value)
}

// Bool returns a raw representation.
func (b Bool) Bool() bool {
	return bool(b)
}

// Kind returns the type of the bool data.
func (b Bool) Kind() Kind {
	return KindBool
}

// Compare compares two Bool values.
func (b Bool) Compare(v Value) int {
	if other, ok := v.(Bool); ok {
		switch {
		case b == other:
			return 0
		case b == TRUE:
			return 1
		default:
			return -1
		}
	}
	if b.Kind() > v.Kind() {
		return 1
	}
	return -1
}

// Interface converts Bool to a bool.
func (b Bool) Interface() any {
	return bool(b)
}

func newBoolEncoder() encoding.Compiler[*Value] {
	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Encoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Bool {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*bool)(target)
					*source = NewBool(t)
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newBoolDecoder() encoding.Compiler[Value] {
	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Encoder[Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Bool {
				return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Bool); ok {
						*(*bool)(target) = s.Bool()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Bool); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
